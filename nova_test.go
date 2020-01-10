package nova

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test adding Routes
func TestServer_All(t *testing.T) {
	msg := "all hit"
	endpoint := "/test"
	s := New()
	s.All(endpoint, func(r *Request) error {
		return r.Send(msg)
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + endpoint)
	if err != nil {
		t.Error(err)
	}

	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if string(data) != msg {
		t.Errorf("All route not hit expected %s got %s", msg, string(data))
	}
}

func TestServer_GetBase(t *testing.T) {
	endpoint := "/"
	s := New()
	s.Get(endpoint, func(r *Request) error {
		return nil
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + endpoint)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("couldn't get 200 from endpoint")
	}
}

func TestServer_GetBase404(t *testing.T) {
	s := New()
	s.Get("/test", func(r *Request) error {
		return nil
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 404 {
		t.Error("couldn't get 200 from endpoint")
	}
}

func TestServer_GetEndpoint(t *testing.T) {
	endpoint := "/test"
	s := New()
	s.Get(endpoint, func(r *Request) error {
		return nil
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + endpoint)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("couldn't get 200 from endpoint")
	}
}

func TestServer_GetParam(t *testing.T) {
	endpoint := "/test/:param"
	s := New()
	s.Get(endpoint, func(r *Request) error {
		if r.RouteParam("param") != "world" {
			t.Error("couldn't get param")
		}
		return nil
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/test/world")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("couldn't get 200 from endpoint")
	}
}

func TestServer_GetPartialRoutes(t *testing.T) {
	endpoint := "/test/:param"
	s := New()
	s.Get(endpoint, func(r *Request) error {
		if r.RouteParam("param") != "world" {
			t.Error("couldn't get param")
		}
		return nil
	})

	s.Get("/test/stuff/hello/", func(req *Request) error {
		return nil
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	//TODO: if route ends with / and route isn't defined with ending / not found
	res, err := http.Get(ts.URL + "/test/stuff/hello")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("couldn't get 200 from endpoint")
	}
}

func TestServer_Put(t *testing.T) {
	endpoint := "/test"
	s := New()
	s.Put(endpoint, func(r *Request) error {
		return nil
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	client := http.Client{}
	req, _ := http.NewRequest(http.MethodPut, ts.URL+endpoint, strings.NewReader("hello"))
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("couldn't make request %s", err)
	}

	if res.StatusCode != 200 {
		t.Error("couldn't get 200 from endpoint")
	}
}

func TestServer_Post(t *testing.T) {
	endpoint := "/test"
	s := New()
	s.Post(endpoint, func(r *Request) error {
		var ts struct {
			Hello string
		}

		err := r.ReadJSON(&ts)
		if err != nil {
			t.Fatal(err)
		}

		if ts.Hello != "world" {
			r.StatusCode(http.StatusBadRequest)
			return r.Send("bad data")
		}

		return nil
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	client := http.Client{}
	req, _ := http.NewRequest(http.MethodPost, ts.URL+endpoint, strings.NewReader(`{"Hello": "world"}`))
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("couldn't make request %s", err)
	}

	if res.StatusCode != 200 {
		t.Error("couldn't get 200 from endpoint")
	}
}

// Check middleware
func TestServer_Use(t *testing.T) {
	s := New()
	s.Use(func(req *Request, next func()) {
		req.Header().Set("Content-Type", "application/json")
	})

	s.Get("/json", func(req *Request) error {
		return nil
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/json")
	if err != nil {
		t.Error(err)
	}

	if res.Header.Get("Content-Type") != "application/json" {
		t.Error("middleware failed Content-Type not set")
	}
}

func TestServer_UseNext(t *testing.T) {
	msg := "json hit"
	endpoint := "/json"
	s := New()
	s.Use(func(req *Request, next func()) {
		req.Header().Set("Content-Type", "application/json")
		next()
	})

	s.Get(endpoint, func(req *Request) error {
		return req.Send(msg)
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + endpoint)
	if err != nil {
		t.Error(err)
	}

	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if res.Header.Get("Content-Type") != "application/json" && string(data) != msg {
		t.Error("middleware failed Content-Type not set")
	}
}

func TestRouteParam(t *testing.T) {
	param := "world"
	endpoint := "/hello/:param"
	s := New()
	s.Get(endpoint, func(r *Request) error {
		return r.Send(r.RouteParam("param"))
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/hello/world")
	if err != nil {
		t.Error(err)
	}

	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if string(data) != param {
		t.Errorf("All route not hit expected %s got %s", param, string(data))
	}
}

func TestQueryParam(t *testing.T) {
	param := "earth"
	endpoint := "/hello/"
	s := New()
	s.Get(endpoint, func(r *Request) error {
		return r.Send(r.QueryParam("world"))
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/hello/?world=earth")
	if err != nil {
		t.Error(err)
	}

	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if string(data) != param {
		t.Errorf("All route not hit expected %s got %s", param, string(data))
	}
}

func TestRequest_JSON(t *testing.T) {
	type holder struct {
		Hello string
	}
	endpoint := "/test"
	s := New()
	s.Get(endpoint, func(r *Request) error {
		ts := holder{
			"world",
		}

		return r.JSON(200, ts)
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + endpoint)
	if err != nil {
		t.Error(err)
	}

	var h holder
	err = json.NewDecoder(res.Body).Decode(&h)
	if err != nil {
		t.Error(err)
	}

	if h.Hello != "world" {
		t.Error("failed to get and parse JSON")
	}
}

func TestRequest_Error(t *testing.T) {
	endpoint := "/test"
	s := New()
	s.Get(endpoint, func(r *Request) error {
		return r.Error(http.StatusNotImplemented, "method not ready", nil)
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + endpoint)
	if err != nil {
		t.Error(err)
	}

	var e JSONErrors
	err = json.NewDecoder(res.Body).Decode(&e)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != http.StatusNotImplemented {
		t.Errorf("got wrong status code expected %d got %d", http.StatusNotImplemented, res.StatusCode)
	}
}

func TestRequest_ErrorCallback(t *testing.T) {
	endpoint := "/test"
	s := New()

	errorHit := false
	s.ErrorFunc(func(req *Request, err error) {
		if err != nil {
			errorHit = true
		}
	})

	s.Get(endpoint, func(r *Request) error {
		return r.Error(http.StatusNotImplemented, "method not ready", errors.New("server error"))
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	_, err := http.Get(ts.URL + endpoint)
	if err != nil {
		t.Error(err)
	}

	if !errorHit {
		t.Fatal("error func wasn't hit when error was returned")
	}
}

func Test404(t *testing.T) {
	endpoint := "/hello/:param"
	s := New()
	s.All(endpoint, func(r *Request) error {
		return r.Send(r.RouteParam("param"))
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/hello/world/more/stuff")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 404 {
		t.Errorf("expected 404 got %d", res.StatusCode)
	}
}

func TestServer_EnableDebug(t *testing.T) {
	s := New()
	s.EnableDebug(true)

	if !s.debug {
		t.Error("Debug mode wasn't set")
	}
}

func TestServer_ErrorFunc(t *testing.T) {
	endpoint := "/test"
	s := New()

	errorHit := false

	s.ErrorFunc(func(req *Request, err error) {
		errorHit = true
	})

	s.Get(endpoint, func(r *Request) error {
		return errors.New("error hit for testing")
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Get(ts.URL + endpoint)
	if err != nil {
		t.Error(err)
	}

	res.Body.Close()

	if !errorHit {
		t.Error("Error function not called")
	}
}

func BenchmarkClimbTree(b *testing.B) {
	m := New()

	m.Get("/test/stuff/world", func(req *Request) error {
		return req.Send("hello")
	})
	for i := 0; i < b.N; i++ {
		m.climbTree("GET", "/test/stuff/world")
	}
}

func TestRouteGroup(t *testing.T) {
	endpoint := "/hello/world"

	m := New()

	g := m.Group("/hello")

	g.Get("/world", func(req *Request) error {
		return nil
	})

	ts := httptest.NewServer(m)
	defer ts.Close()

	res, err := http.Get(ts.URL + endpoint)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("couldn't get 200 from endpoint")
	}
}
