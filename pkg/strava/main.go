package strava

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	strava "github.com/strava/go.strava"
	"golang.org/x/oauth2"
)

// transport is an http.RoundTripper that keeps track of the in-flight
// request and implements hooks to report HTTP tracing events.
type transport struct {
	tp      *http.Transport
	current *http.Request
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.WithValue(req.Context(), "RequestStart", time.Now())
	req = req.WithContext(ctx)
	resp, err := t.tp.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if start, ok := ctx.Value("RequestStart").(time.Time); ok {
		log.Printf("%s %d %s (%s)", req.Method, resp.StatusCode, resp.Request.URL, time.Since(start))
	}

	return resp, err
}

func GetClient(token *oauth2.Token) *strava.Client {
	t := &transport{
		tp: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			//ExpectContinueTimeout: 1 * time.Second,
		}}
	return strava.NewClient(token.AccessToken, &http.Client{Transport: t})
}
