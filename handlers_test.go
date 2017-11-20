package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPosHandler(t *testing.T) {
	tests := []struct {
		name                                  string
		j                                     string
		ifaValue                              string
		appValue, platformValue, countryValue string
		pos                                   int
		response                              string
	}{
		// TODO: Add test cases.
		{"base",
			`{"id":"CA1DF4146DE67248","imp":[{"id":"1","banner":{"w":320,"h":480,"pos":1,"btype":[4],"battr":[3,8,10,14],"api":[3,5]},"video":{"mimes":["video/3gpp","video/mp4"],"minduration":16,"maxduration":120,"protocols":[2,5,3,6],"w":320,"h":480,"linearity":1,"sequence":1,"battr":[3,8,10,14],"api":[3,5],"companiontype":[1,2,3]},"displaymanager":"mopub","displaymanagerver":"4.7.1","instl":1,"tagid":"01a16f7530db4b6689d10d7d5cab3183","bidfloor":25.22,"ext":{"brsrclk":1,"dlp":1}}],"app":{"bundle":"com.erenapps.beachatvsim","ver":"29","id":"7271c98bfac340ff83c0bf3008f31d55","name":"Beach ATV Simulator - 16478","cat":["IAB1","IAB9","IAB9-30","entertainment","games"],"publisher":{"id":"57e33763e5804d75802705596150f3c1","name":"Appodeal, Inc."}},"device":{"ua":"Mozilla/5.0 (Linux; Android 4.4.4; SM-T561 Build/KTU84P) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/33.0.0.0 Safari/537.36","ip":"178.206.218.71","geo":{"lat":54.55,"lon":52.8,"accuracy":50,"ipservice":3,"country":"RUS","region":"73","city":"Bugulma","zip":"423230","type":2,"utcoffset":180,"ext":{"old_geo":{"country":"RUS","region":"73","city":"Bugulma","zip":"423230"}}},"carrier":"250-02","language":"ru","make":"samsung","model":"SM-T561","os":"Android","osv":"4.4.4","js":1,"connectiontype":2,"ifa":"951bfcee-3fc4-45c3-82b2-50371681e28a","h":800,"w":1280,"pxratio":1},"at":2,"cur":["USD"],"bcat":["IAB25","IAB26","IAB9-9"],"ext":{"envisionx":{"ssp":9}}}`,
			"951bfcee-3fc4-45c3-82b2-50371681e28a",
			"com.erenapps.beachatvsim", "Android", "RUS",
			1,
			"{\"pos\": 1}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := PosHandler(
				posCounterFunc(func(ifa string) (int, error) {
					if ifa != tt.ifaValue {
						t.Errorf("Wrong ifa value: %s (want: %s)", ifa, tt.ifaValue)
					}
					return tt.pos, nil
				}),
				statsCollectorFunc(func(app, platform, country string) error {
					if app != tt.appValue {
						t.Errorf("Wrong app value: %s (want: %s)", app, tt.appValue)
					}
					if platform != tt.platformValue {
						t.Errorf("Wrong platform value: %s (want: %s)", platform, tt.platformValue)
					}
					if country != tt.countryValue {
						t.Errorf("Wrong country value: %s (want: %s)", country, tt.countryValue)
					}
					return nil
				}),
			)

			r, err := http.NewRequest("GET", "/", strings.NewReader(tt.j))
			if err != nil {
				t.Fatalf("Error on create request: %v", err)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)

			if w.Code != 200 {
				t.Errorf("Wrong response code: %d", w.Code)
			}
			if w.Header().Get("Content-Type") != "application/json" {
				t.Errorf("Wrong response content-type: %v", w.Header())
			}
			if w.Body.String() != tt.response {
				t.Errorf("Want response body:  %s", tt.response)
				t.Errorf("Wrong response body: %s", w.Body.String())
			}
		})
	}
}

func TestPosHandler_Errors(t *testing.T) {
	handler := PosHandler(
		posCounterFunc(func(ifa string) (int, error) {
			return 0, errors.New("test error")
		}),
		statsCollectorFunc(func(app, platform, country string) error {
			return errors.New("test 2 error")
		}),
	)

	// bad request
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Error on create request: %v", err)
	}
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 400 {
		t.Errorf("Wrong status: %d", w.Code)
	}
	if w.Body.String() != "Bad request: 'body' is empty\n" {
		t.Errorf("Wrong error message: %q", w.Body.String())
	}

	// bad request - empty ifa
	r, err = http.NewRequest("GET", "/", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("Error on create request: %v", err)
	}
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 400 {
		t.Errorf("Wrong status: %d", w.Code)
	}
	if w.Body.String() != "Bad request: 'ifa' is empty\n" {
		t.Errorf("Wrong error message: %q", w.Body.String())
	}

	// bad pos
	r, err = http.NewRequest("GET", "/", strings.NewReader(`{"device": {"ifa": "test"}}`))
	if err != nil {
		t.Fatalf("Error on create request: %v", err)
	}
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != 500 {
		t.Errorf("Wrong status: %d", w.Code)
	}
	if w.Body.String() != "Error on get pos: test error\n" {
		t.Errorf("Wrong response body: %s", w.Body.String())
	}

	// bad stats
	handler = PosHandler(
		posCounterFunc(func(ifa string) (int, error) {
			return 1, nil
		}),
		statsCollectorFunc(func(app, platform, country string) error {
			return errors.New("test 2 error")
		}),
	)
	r, err = http.NewRequest("GET", "/", strings.NewReader(`{"device": {"ifa": "test"}}`))
	if err != nil {
		t.Fatalf("Error on create request: %v", err)
	}
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != 500 {
		t.Errorf("Wrong status: %d", w.Code)
	}
	if w.Body.String() != "Error on add stats: test 2 error\n" {
		t.Errorf("Wrong response body: %s", w.Body.String())
	}
}

type posCounterFunc func(ifa string) (int, error)

func (f posCounterFunc) Pos(ifa string) (int, error) { return f(ifa) }

type statsCollectorFunc func(app, platform, country string) error

func (f statsCollectorFunc) Add(app, platform, country string) error { return f(app, platform, country) }

func TestStatsHandler(t *testing.T) {
	handler := StatsHandler(statsGetterFunc(func() ([]*Stat, error) {
		return []*Stat{{App: "12345", Platform: "ios", Country: "RUS", Count: 500}}, nil
	}))

	r, err := http.NewRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatalf("Error on create request: %v", err)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("Wrong response status: %d", w.Code)
	}
	if w.Body.String() != "[{\"app\":\"12345\",\"platform\":\"ios\",\"country\":\"RUS\",\"count\":500}]\n" {
		t.Errorf("Wrong response body: %q", w.Body.String())
	}

	// bad stats
	handler = StatsHandler(statsGetterFunc(func() ([]*Stat, error) {
		return nil, errors.New("test error")
	}))
	r, err = http.NewRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatalf("Error on create request: %v", err)
	}
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != 500 {
		t.Errorf("Wrong status: %d", w.Code)
	}
	if w.Body.String() != "Error on get stats: test error\n" {
		t.Errorf("Wrong response body: %q", w.Body.String())
	}
}

type statsGetterFunc func() ([]*Stat, error)

func (f statsGetterFunc) GetAll() ([]*Stat, error) { return f() }
