package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	apiclient "github.com/Mattias-/strava-go/client"
	activitiesapi "github.com/Mattias-/strava-go/client/activities"
	stravamodels "github.com/Mattias-/strava-go/models"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	geojson "github.com/paulmach/go.geojson"
	"golang.org/x/oauth2"

	"github.com/Mattias-/strava-multi-activity-mapper/pkg/queryparser"
)

var (
	commit  = "none"
	date    = "unknown"
	baseUrl string
	port    string
	conf    *oauth2.Config
)

func main() {
	clientID, ok := os.LookupEnv("CLIENT_ID")
	if !ok {
		log.Fatalln("env CLIENT_ID not set")
	}
	clientSecret, ok := os.LookupEnv("CLIENT_SECRET")
	if !ok {
		log.Fatalln("env CLIENT_SECRET not set")
	}
	cookieSecret, ok := os.LookupEnv("COOKIE_SECRET")
	if !ok {
		log.Fatalln("env COOKIE_SECRET not set")
	}
	port, ok = os.LookupEnv("PORT")
	if !ok {
		port = "8000"
	}
	baseUrl, ok = os.LookupEnv("BASE_URL")
	if !ok {
		baseUrl = "http://localhost:" + port
	}
	conf = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"activity:read_all"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.strava.com/oauth/authorize",
			TokenURL: "https://www.strava.com/oauth/token",
		},
		RedirectURL: baseUrl + "/callback",
	}

	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(cookieSecret))))
	e.Use(middleware.Logger())

	e.Static("/", "dist")
	e.GET("/version", version)
	e.GET("/auth", auth)
	e.GET("/callback", callback)
	e.GET("/athlete", athlete, withToken)
	e.GET("/activities", activities, withToken)
	e.GET("/activitytypes", activityTypes)

	e.Logger.Fatal(e.Start(":" + port))
}

func version(c echo.Context) error {
	r := struct {
		Commit string `json:"commit"`
		Date   string `json:"date"`
	}{
		Commit: commit,
		Date:   date,
	}
	return c.JSON(http.StatusOK, r)
}

func auth(c echo.Context) error {
	sess, _ := session.Get("session", c)
	state := uuid.New().String()
	sess.Values["oauth-state"] = state
	err := sess.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}
	authUrl := conf.AuthCodeURL(state, oauth2.SetAuthURLParam("show_dialog", "true"))
	return c.Redirect(http.StatusTemporaryRedirect, authUrl)
}

func callback(c echo.Context) error {
	sess, _ := session.Get("session", c)
	state := sess.Values["oauth-state"]
	if state != c.QueryParam("state") {
		return c.String(http.StatusBadRequest, "Invalid state")
	}
	ctx := context.Background()
	tok, err := conf.Exchange(ctx, c.QueryParam("code"))
	if err != nil {
		return err
	}

	sess.Values["oauth-accesstoken"] = tok.AccessToken
	sess.Values["oauth-refreshtoken"] = tok.RefreshToken
	sess.Values["oauth-expiry"] = tok.Expiry.Unix()
	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusTemporaryRedirect, baseUrl)
}

func getToken(c echo.Context) (*oauth2.Token, error) {
	sess, _ := session.Get("session", c)
	accessToken, ok := sess.Values["oauth-accesstoken"].(string)
	if !ok {
		return nil, c.JSON(http.StatusBadRequest, "Missing oauth token")
	}
	refreshToken, ok := sess.Values["oauth-refreshtoken"].(string)
	if !ok {
		return nil, c.JSON(http.StatusBadRequest, "Missing oauth token")
	}
	expiry, ok := sess.Values["oauth-expiry"].(int64)
	if !ok {
		return nil, c.JSON(http.StatusBadRequest, "Missing oauth token")
	}
	c.Logger().Debug(sess.Values)

	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       time.Unix(expiry, 0),
	}
	newToken, err := conf.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, err
	}
	sess.Values["oauth-accesstoken"] = newToken.AccessToken
	sess.Values["oauth-refreshtoken"] = newToken.RefreshToken
	sess.Values["oauth-expiry"] = newToken.Expiry.Unix()
	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return nil, err
	}
	return newToken, nil
}

func withToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := getToken(c)
		if err != nil {
			return err
		}
		c.Set("token", token)
		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}

func athlete(c echo.Context) error {
	token := c.Get("token").(*oauth2.Token)
	r := httptransport.New(apiclient.DefaultHost, apiclient.DefaultBasePath, apiclient.DefaultSchemes)
	r.DefaultAuthentication = httptransport.BearerToken(token.AccessToken)
	client := apiclient.New(r, strfmt.Default)

	athlete, err := client.Athletes.GetLoggedInAthlete(nil, nil)
	if err != nil {
		fmt.Printf("%w", err)
		return err
	}
	return c.JSON(http.StatusOK, athlete.Payload)
}

func activityTypes(c echo.Context) error {
	return c.JSON(http.StatusOK, nil)
}

func roundUp(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}

func roundDown(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 1, 0, t.Location())
}

func ptrint64(p int64) *int64 {
	return &p
}

func getActivities(client *apiclient.StravaAPIV3, beforeS string, afterS string, outChan chan *stravamodels.SummaryActivity) {
	before, err := time.Parse("2006-01-02", beforeS)
	if beforeS == "" || err != nil {
		before = time.Now()
	}
	before = roundUp(before)

	after, err := time.Parse("2006-01-02", afterS)
	if afterS == "" || err != nil {
		after = time.Now()
	}
	after = roundDown(after)

	perPage := 200
	params := activitiesapi.NewGetLoggedInAthleteActivitiesParams().
		WithPerPage(ptrint64(int64(perPage))).
		WithBefore(ptrint64(before.Unix())).
		WithAfter(ptrint64(after.Unix()))

	for i := 1; ; i++ {
		ac, err := client.Activities.GetLoggedInAthleteActivities(params.WithPage(ptrint64(int64(i))), nil)
		if err != nil {
			log.Println(err)
		}
		a := ac.Payload
		log.Printf("Got %d activities", len(a))
		for _, activity := range a {
			outChan <- activity
		}
		if len(a) != perPage {
			break
		}
	}
	close(outChan)
}

func activityFeature(client *apiclient.StravaAPIV3, query string, activityType string, activity *stravamodels.SummaryActivity, featureChan chan *geojson.Feature) {
	if activityType != "" && activityType != string(activity.Type) {
		// Don't add activies that doesn't match the type filter.
		return
	}
	if queryparser.Matches(activity.Name, query) {
		// We found a match in the activity name
		featureChan <- activityToGeoJSON(activity)
	} else {
		params := activitiesapi.NewGetActivityByIDParams()
		activity2, err := client.Activities.GetActivityByID(params.WithID(int64(activity.ID)), nil)
		if err == nil {
			if queryparser.Matches(activity2.Payload.Description, query) {
				// We found a match in the description
				featureChan <- activityToGeoJSON(&activity2.Payload.SummaryActivity)
			}
		} else {
			log.Printf("Could not get activity %d: %v", activity.ID, err)
		}
	}
}

func athleteFeatures(client *apiclient.StravaAPIV3, before, after, q, at string) chan *geojson.Feature {
	activityChan := make(chan *stravamodels.SummaryActivity)
	go getActivities(client, before, after, activityChan)

	var wg sync.WaitGroup
	featureChan := make(chan *geojson.Feature)
	for activity := range activityChan {
		wg.Add(1)
		go func(activity *stravamodels.SummaryActivity) {
			activityFeature(client, q, at, activity, featureChan)
			wg.Done()
		}(activity)
	}
	go func() {
		// Wait for all features to be created.
		wg.Wait()
		// At that point we can close the channel.
		close(featureChan)
	}()
	return featureChan
}

func activities(c echo.Context) error {
	token := c.Get("token").(*oauth2.Token)
	r := httptransport.New(apiclient.DefaultHost, apiclient.DefaultBasePath, apiclient.DefaultSchemes)
	r.DefaultAuthentication = httptransport.BearerToken(token.AccessToken)
	client := apiclient.New(r, strfmt.Default)

	featureChan := athleteFeatures(client,
		c.QueryParam("before"),
		c.QueryParam("after"),
		c.QueryParam("q"),
		c.QueryParam("type"),
	)
	fc := geojson.NewFeatureCollection()
	for a := range featureChan {
		fc.AddFeature(a)
	}

	return c.JSON(http.StatusOK, fc)
}

func activityToGeoJSON(as *stravamodels.SummaryActivity) *geojson.Feature {
	var p string
	if as.Map.Polyline != "" {
		p = as.Map.Polyline
	} else {
		p = as.Map.SummaryPolyline
	}
	vsm := decodePoly(p)
	f := geojson.NewLineStringFeature(vsm)
	f.Properties["name"] = as.Name
	f.Properties["type"] = as.Type
	f.Properties["id"] = as.ID
	f.Properties["start_date_local"] = as.StartDateLocal
	f.Properties["activity"] = as
	return f
}

func decodePoly(p string) [][]float64 {
	var count, index int
	factor := 1.0e5

	line := make([][]float64, 0)
	tempLatLng := [2]int{0, 0}

	for index < len(p) {
		var result int
		var b int = 0x20
		var shift uint

		for b >= 0x20 {
			b = int(p[index]) - 63
			index++

			result |= (b & 0x1f) << shift
			shift += 5
		}

		// sign dection
		if result&1 != 0 {
			result = ^(result >> 1)
		} else {
			result = result >> 1
		}

		if count%2 == 0 {
			result += tempLatLng[0]
			tempLatLng[0] = result
		} else {
			result += tempLatLng[1]
			tempLatLng[1] = result

			line = append(line, []float64{float64(tempLatLng[1]) / factor, float64(tempLatLng[0]) / factor})
		}

		count++
	}

	return line
}
