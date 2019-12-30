package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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

func user(client *http.Client) {
	res, err := client.Get("https://www.strava.com/api/v3/athlete")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(body))
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

func roundUp(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}

func roundDown(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 1, 0, t.Location())
}

func activity(c echo.Context) error {
	token, err := getToken(c)
	if err != nil {
		return err
	}

	client := strava.NewClient(token.AccessToken)
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

	// TODO: Paginate until after date.
	activities, err := ca.ListActivities().
		PerPage(200).
		Before(int(before.Unix())).
		After(int(after.Unix())).
		Do()

	if err != nil {
		return err
	}
	fc := geojson.NewFeatureCollection()
	for _, activity := range activities {
		if strings.Contains(activity.Name, c.QueryParam("q")) {
			fc.AddFeature(activityToGeoJSON(activity))
			continue
		}
		activity2, err := as.Get(activity.Id).Do()
		if err == nil {
			if strings.Contains(activity2.Description, c.QueryParam("q")) {
				fc.AddFeature(activityToGeoJSON(activity))
				continue
			}
		} else {
			log.Printf("Could not get activity %d: %v", activity.Id, err)
		}
	}
	return c.JSON(http.StatusOK, fc)
}

func activityToGeoJSON(as *strava.ActivitySummary) *geojson.Feature {
	vs := as.Map.SummaryPolyline.Decode()
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
