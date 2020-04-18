package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/paulmach/go.geojson"
	"github.com/strava/go.strava"
	"golang.org/x/oauth2"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
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
	//e.Use(middleware.Recover())

	e.File("/", "index.html")
	e.Static("/static", "static")
	e.GET("/auth", auth)
	e.GET("/activity", activity)
	e.GET("/athlete", athlete)
	e.GET("/callback", callback)
	e.GET("/activitytypes", activityTypes)

	e.Logger.Fatal(e.Start(":" + port))
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

func athlete(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return err
	}

	client := strava.NewClient(token.AccessToken)
	service := strava.NewCurrentAthleteService(client)

	athlete, err := service.Get().Do()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, athlete)
}

func activityTypes(c echo.Context) error {
	return c.JSON(http.StatusOK, strava.ActivityTypes)
}

func roundUp(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}

func roundDown(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 1, 0, t.Location())
}

// transport is an http.RoundTripper that keeps track of the in-flight
// request and implements hooks to report HTTP tracing events.
type transport struct {
	current *http.Request
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.WithValue(req.Context(), "RequestStart", time.Now())
	req = req.WithContext(ctx)

	log.Printf("--> %s %s", req.Method, req.URL)

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if start, ok := ctx.Value("RequestStart").(time.Time); ok {
		log.Printf("<-- %d %s (%s)", resp.StatusCode, resp.Request.URL, time.Now().Sub(start))
	}

	return resp, err
}

func activity(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return err
	}

	t := &transport{}
	client := strava.NewClient(token.AccessToken, &http.Client{Transport: t})
	ca := strava.NewCurrentAthleteService(client)
	as := strava.NewActivitiesService(client)

	before, err := time.Parse("2006-01-02", c.QueryParam("before"))
	if c.QueryParam("before") == "" || err != nil {
		before = time.Now()
	}

	after, err := time.Parse("2006-01-02", c.QueryParam("after"))
	if c.QueryParam("after") == "" || err != nil {
		after = time.Now()
	}
	before = roundUp(before)
	after = roundDown(after)

	perPage := 200
	activities := []*strava.ActivitySummary{}
	for i := 1; ; i++ {
		a, err := ca.ListActivities().
			Page(i).
			PerPage(perPage).
			Before(int(before.Unix())).
			After(int(after.Unix())).
			Do()
		if err != nil {
			log.Println(err)
			return err
		}
		activities = append(activities, a...)
		log.Printf("Got %d activities", len(a))
		if len(a) != perPage {
			break
		}
	}

	fc := geojson.NewFeatureCollection()
	var wg sync.WaitGroup
	wg.Add(len(activities))
	for _, activity := range activities {
		go func(activity *strava.ActivitySummary) {
			defer wg.Done()
			if c.QueryParam("type") != "" && c.QueryParam("type") != string(activity.Type) {
				// Don't add activies that doesn't match the type filter.
				return
			}
			if strings.Contains(activity.Name, c.QueryParam("q")) {
				// We found a match in the activity name
				fc.AddFeature(activityToGeoJSON(activity))
				return
			}
			activity2, err := as.Get(activity.Id).Do()
			if err == nil {
				if strings.Contains(activity2.Description, c.QueryParam("q")) {
					// We found a match in the description
					fc.AddFeature(activityToGeoJSON(&activity2.ActivitySummary))
					return
				}
			} else {
				log.Printf("Could not get activity %d: %v", activity.Id, err)
			}
		}(activity)
	}
	wg.Wait()
	sort.Slice(fc.Features, func(i, j int) bool {
		t1 := (*fc.Features[i]).Properties["start_date_local"].(time.Time)
		t2 := (*fc.Features[j]).Properties["start_date_local"].(time.Time)
		// reverse chronological order
		return t2.Before(t1)
	})
	return c.JSON(http.StatusOK, fc)
}

func activityToGeoJSON(as *strava.ActivitySummary) *geojson.Feature {
	var p strava.Polyline
	if as.Map.Polyline != "" {
		p = as.Map.Polyline
	} else {
		p = as.Map.SummaryPolyline
	}
	vs := p.Decode()
	vsm := make([][]float64, len(vs))
	for i, v := range vs {
		vsm[i] = []float64{v[1], v[0]}
	}
	f := geojson.NewLineStringFeature(vsm)
	f.Properties["name"] = as.Name
	f.Properties["type"] = as.Type
	f.Properties["id"] = as.Id
	f.Properties["start_date_local"] = as.StartDateLocal
	f.Properties["activity"] = as
	return f
}
