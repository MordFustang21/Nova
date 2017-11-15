package nova

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"io/ioutil"
	"strings"
	"encoding/json"
)

// Test adding Routes
func TestServer_All(t *testing.T) {
	msg := "all hit"
	endpoint := "/test/"
	s := New()
	s.All(endpoint, func(r *Request) {
		r.Send(msg)
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

func TestServer_Get(t *testing.T) {
	endpoint := "/test"
	s := New()
	s.Get(endpoint, func(r *Request) {
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

func TestServer_Put(t *testing.T) {
	endpoint := "/test"
	s := New()
	s.Put(endpoint, func(r *Request) {
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	client := http.Client{}
	req, _ := http.NewRequest(http.MethodPut, ts.URL + endpoint, strings.NewReader("hello"))
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
	s.Post(endpoint, func(r *Request) {
		var ts struct {
			Hello string
		}

		r.ReadJSON(&ts)

		if ts.Hello != "world" {
			r.StatusCode(http.StatusBadRequest)
			r.Send("bad data")
		}
	})

	ts := httptest.NewServer(s)
	defer ts.Close()

	client := http.Client{}
	req, _ := http.NewRequest(http.MethodPost, ts.URL + endpoint, strings.NewReader(`{"Hello": "world"}`))
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("couldn't make request %s", err)
	}

	if res.StatusCode != 200 {
		t.Error("couldn't get 200 from endpoint")
	}
}
func TestServer_Delete(t *testing.T) {
	s := New()
	s.Delete("/test", func(r *Request) {

	})

	if s.paths["DELETE"].children["test"] == nil {
		t.Error("Failed to insert DELETE route")
	}
}

// Check middleware
func TestServer_Use(t *testing.T) {
	s := New()
	s.Use(func(req *Request, next func()) {
		req.Response.Header().Set("Content-Type", "application/json")
	})

	s.Get("/json", func(req *Request) {

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
		req.Response.Header().Set("Content-Type", "application/json")
		next()
	})

	s.Get(endpoint, func(req *Request) {
		req.Send(msg)
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

func TestServer_Restricted(t *testing.T) {
	s := New()
	s.Restricted("OPTION", "/test", func(*Request) {

	})

	if s.paths["OPTION"].children["test"] == nil {
		t.Error("Route wasn't restricted to method")
	}
}

func TestMultipleChildren(t *testing.T) {
	s := New()
	s.All("/test/stuff", func(*Request) {

	})

	s.All("/test/test", func(*Request) {

	})

	if len(s.paths[""].children["test"].children) != 2 {
		t.Error("Node possibly overwritten")
	}
}

func TestRouteParam(t *testing.T) {
	param := "world"
	endpoint := "/hello/:param"
	s := New()
	s.Get(endpoint, func(r *Request) {
		r.Send(r.RouteParam("param"))
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
	s.Get(endpoint, func(r *Request) {
		r.Send(r.QueryParam("world"))
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
	s.Get(endpoint, func(r *Request) {
		ts := holder{
			"world",
		}

		r.JSON(200, ts)
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
	s.Get(endpoint, func(r *Request) {
		r.Error(http.StatusNotImplemented, "method not ready")
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

func Test404(t *testing.T) {
	endpoint := "/hello/:param"
	s := New()
	s.All(endpoint, func(r *Request) {
		r.Send(r.RouteParam("param"))
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

// Test finding Routes
func TestServer_climbTree(t *testing.T) {
	cases := []struct {
		Method    string
		Path      string
		ExpectNil bool
	}{
		{
			"GET",
			"/test",
			false,
		},
		{
			"GET",
			"/stuff/param1/params/param2/",
			false,
		},
		{
			"GET",
			"/stuff/param1/par/param2",
			true,
		},
	}

	s := New()
	s.Get("/test", func(*Request) {

	})

	s.Get("/stuff/:test/params/:more", func(*Request) {

	})

	for _, val := range cases {
		node := s.climbTree(val.Method, val.Path)
		if val.ExpectNil && node != nil {
			t.Errorf("%s Expected nil got *Node", val.Path)
		} else if !val.ExpectNil && node == nil {
			t.Errorf("%s Expected *Node got nil", val.Path)
		}
	}
}

func TestServer_EnableDebug(t *testing.T) {
	s := New()
	s.EnableDebug(true)

	if !s.debug {
		t.Error("Debug mode wasn't set")
	}
}
