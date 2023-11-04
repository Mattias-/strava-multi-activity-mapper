package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"sort"
	"strings"
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
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

var (
	baseUrl string
	port    string
	conf    *oauth2.Config
)

func main() {
	h := slog.NewJSONHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(h))
	clientID, ok := os.LookupEnv("CLIENT_ID")
	if !ok {
		slog.Error("env CLIENT_ID not set")
		os.Exit(1)
	}
	clientSecret, ok := os.LookupEnv("CLIENT_SECRET")
	if !ok {
		slog.Error("env CLIENT_SECRET not set")
		os.Exit(1)
	}
	cookieSecret, ok := os.LookupEnv("COOKIE_SECRET")
	if !ok {
		slog.Error("env COOKIE_SECRET not set")
		os.Exit(1)
	}
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
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
	cs := sessions.NewCookieStore([]byte(cookieSecret))
	if cookieDomain != "" {
		cs.Options.Domain = cookieDomain
		cs.Options.SameSite = http.SameSiteNoneMode
		cs.Options.Secure = true
	}
	e.Use(session.Middleware(cs))
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{baseUrl, "http://localhost:8000"},
		AllowCredentials: true,
	}))

	e.Static("/", "frontend/dist")
	e.GET("/version", version)
	e.GET("/auth", auth)
	e.GET("/callback", callback)
	e.GET("/athlete", athlete, withToken)
	e.GET("/activities", activities, withToken)
	e.GET("/activitytypes", activityTypes)

	e.Logger.Fatal(e.Start(":" + port))
}

func version(c echo.Context) error {
	commit := "none"
	date := "unknown"
	dirty := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, kv := range info.Settings {
			if kv.Key == "vcs.revision" {
				commit = kv.Value
			}
			if kv.Key == "vcs.time" {
				date = kv.Value
			}
			if kv.Key == "vcs.modified" {
				if kv.Value == "true" {
					dirty = "-dirty"
				}
			}
		}
	}
	r := struct {
		Commit string `json:"commit"`
		Date   string `json:"date"`
	}{
		Commit: commit + dirty,
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
	sess, err := session.Get("session", c)
	if err != nil {
		return nil, c.JSON(http.StatusUnauthorized, nil)
	}
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
	slog.Debug("Got token from session", "token", sess.Values)

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

type problemResponse struct {
	Title  string `json:"title"`
	Status string `json:"status"`
}

func athlete(c echo.Context) error {
	token, ok := c.Get("token").(*oauth2.Token)
	if !ok || token == nil {
		return c.JSON(http.StatusUnauthorized, struct{ status int }{
			status: http.StatusUnauthorized,
		})
	}
	r := httptransport.New(apiclient.DefaultHost, apiclient.DefaultBasePath, apiclient.DefaultSchemes)
	r.DefaultAuthentication = httptransport.BearerToken(token.AccessToken)
	client := apiclient.New(r, strfmt.Default)

	athlete, err := client.Athletes.GetLoggedInAthlete(nil, nil)
	if err != nil {
		slog.Error("Could not get athlete", "error", err.Error())
		if strings.Contains(err.Error(), "Rate Limit Exceeded") {
			return c.JSON(http.StatusTooManyRequests, problemResponse{Title: "Rate Limit Exceeded", Status: fmt.Sprintf("%d", http.StatusTooManyRequests)})
		}
		return err
	}
	return c.JSON(http.StatusOK, athlete.Payload)
}

func activityTypes(c echo.Context) error {
	ats := []byte(`{
  "Ride": "Ride",
  "Run": "Run",
  "Swim": "Swim",
  "Hike": "Hike",
  "Walk": "Walk",
  "AlpineSki": "Alpine Ski",
  "BackcountrySki": "Backcountry Ski",
  "Canoeing": "Canoeing",
  "Crossfit": "Crossfit",
  "EBikeRide": "E-Bike Ride",
  "Elliptical": "Elliptical",
  "Handcycle": "Handcycle",
  "IceSkate": "Ice Skate",
  "InlineSkate": "Inline Skate",
  "Kayaking": "Kayaking",
  "Kitesurf": "Kitesurf",
  "NordicSki": "Nordic Ski",
  "RockClimbing": "Rock Climbing",
  "RollerSki": "Roller Ski",
  "Rowing": "Rowing",
  "Snowboard": "Snowboard",
  "Snowshoe": "Snowshoe",
  "StairStepper": "Stair-Stepper",
  "StandUpPaddling": "Stand Up Paddling",
  "Surfing": "Surfing",
  "Velomobile": "Velomobile",
  "VirtualRide": "Virtual Ride",
  "VirtualRun": "Virtual Run",
  "WeightTraining": "Weight Training",
  "Wheelchair": "Wheelchair",
  "Windsurf": "Windsurf",
  "Workout": "Workout",
  "Yoga": "Yoga"
}`)
	return c.JSONBlob(http.StatusOK, ats)
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
			slog.Error("athlete activites", "error", err.Error())
		}
		a := ac.Payload
		for _, activity := range a {
			outChan <- activity
		}
		if len(a) != perPage {
			break
		}
	}
	close(outChan)
}

func detailedActivity(client *apiclient.StravaAPIV3, query string, activityType string, activity *stravamodels.SummaryActivity, activityChan chan *stravamodels.DetailedActivity) {
	if activityType != "" && activityType != string(activity.Type) {

		slog.Info("Non matching activity", "filter_activity_type", activityType, "acitivity_type", activity.Type)
		// Don't add activies that doesn't match the type filter.
		return
	}
	params := activitiesapi.NewGetActivityByIDParams()
	activity2, err := client.Activities.GetActivityByID(params.WithID(int64(activity.ID)), nil)
	if err != nil {
		slog.Error("Could not get detailed activity", "id", activity.ID, "error", err.Error())
		return
	}

	if query == "" || Matches(activity.Name, query) || Matches(activity2.Payload.Description, query) {
		activityChan <- activity2.Payload
	} else {
		slog.Info("Activity did not match query",
			"query", query,
			"activity_name", activity.Name,
			"activity_description", activity2.Payload.Description,
		)
	}
}

func athleteActivities(client *apiclient.StravaAPIV3, before, after, q, at string) chan *stravamodels.DetailedActivity {
	activityChan := make(chan *stravamodels.SummaryActivity)
	go getActivities(client, before, after, activityChan)

	var wg sync.WaitGroup
	c := make(chan *stravamodels.DetailedActivity)
	for activity := range activityChan {
		wg.Add(1)
		go func(activity *stravamodels.SummaryActivity) {
			detailedActivity(client, q, at, activity, c)
			wg.Done()
		}(activity)
	}
	go func() {
		// Wait for all features to be created.
		wg.Wait()
		// At that point we can close the channel.
		close(c)
	}()
	return c
}

func activities(c echo.Context) error {
	token := c.Get("token").(*oauth2.Token)
	r := httptransport.New(apiclient.DefaultHost, apiclient.DefaultBasePath, apiclient.DefaultSchemes)
	r.DefaultAuthentication = httptransport.BearerToken(token.AccessToken)
	client := apiclient.New(r, strfmt.Default)

	activityChan := athleteActivities(client,
		c.QueryParam("before"),
		c.QueryParam("after"),
		c.QueryParam("q"),
		c.QueryParam("type"),
	)
	fc := geojson.NewFeatureCollection()
	for a := range activityChan {
		f := activityToGeoJSON(a)
		fc.AddFeature(f)
	}

	return c.JSON(http.StatusOK, fc)
}

func activityToGeoJSON(activity *stravamodels.DetailedActivity) *geojson.Feature {
	var p string
	if activity.Map.Polyline != "" {
		p = activity.Map.Polyline
	} else {
		p = activity.Map.SummaryPolyline
	}
	vsm := decodePoly(p)
	f := geojson.NewLineStringFeature(vsm)
	f.Properties["name"] = activity.Name
	f.Properties["description"] = activity.Description
	f.Properties["type"] = activity.Type
	f.Properties["id"] = activity.ID
	f.Properties["start_date_local"] = activity.StartDateLocal
	f.Properties["activity"] = activity
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

func Matches(data, query string) bool {
	dataWords := strings.Split(data, " ")
	queryWords := strings.Split(query, " ")
	sort.Strings(dataWords)
	for _, qw := range queryWords {
		if qw == "" {
			continue
		}
		if contains(dataWords, qw) {
			return true
		}
	}
	return false
}

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}
