package nova

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"
)

// Request resembles an incoming request
type Request struct {
	*http.Request
	Response    *Response
	routeParams map[string]string
	urlValues   url.Values
	BaseUrl     string
	Ctx         context.Context
}

// JSONError resembles the RESTful standard for an error response
type JSONError struct {
	Errors  []interface{} `json:"errors"`
	Code    int           `json:"code"`
	Message string        `json:"message"`
}

// JSONErrors holds the JSONError response
type JSONErrors struct {
	Error JSONError `json:"error"`
}

// NewRequest creates a new Request pointer for an incoming request
func NewRequest(w http.ResponseWriter, r *http.Request) *Request {
	req := new(Request)
	req.Request = r
	req.Response = &Response{w, 200}
	req.routeParams = make(map[string]string)
	req.urlValues = r.URL.Query()
	req.BaseUrl = r.RequestURI

	return req
}

// RouteParam checks for and returns param or "" if doesn't exist
func (r *Request) RouteParam(key string) string {
	if val, ok := r.routeParams[key]; ok {
		return val
	}

	return ""
}

// QueryParam checks for and returns param or "" if doesn't exist
func (r *Request) QueryParam(key string) string {
	return r.urlValues.Get(key)
}

// Error allows an easy method to set the RESTful standard error response
func (r *Request) Error(statusCode int, msg string, errors ...interface{}) error {
	newErr := JSONErrors{
		Error: JSONError{
			Errors:  errors,
			Code:    statusCode,
			Message: msg,
		},
	}
	return r.JSON(statusCode, newErr)
}

// buildRouteParams builds a map of the route params
func (r *Request) buildRouteParams(route string) {
	routeParams := r.routeParams
	reqParts := strings.Split(r.BaseUrl[1:], "/")
	routeParts := strings.Split(route[1:], "/")

	for index, val := range routeParts {
		if val[0] == ':' {
			routeParams[val[1:]] = reqParts[index]
		}
	}
}

// ReadJSON unmarshals request body into the struct provided
func (r *Request) ReadJSON(i interface{}) error {
	return json.NewDecoder(r.Request.Body).Decode(i)
}

// Send writes the data to the response body
func (r *Request) Send(data interface{}) error {
	var err error

	switch v := data.(type) {
	case []byte:
		_, err = r.Response.Write(v)
	case string:
		_, err = r.Response.Write([]byte(v))
	case error:
		_, err = r.Response.Write([]byte(v.Error()))
	default:
		err = errors.New("unsupported type Send type")
	}

	return err
}

// JSON marshals the given interface object and writes the JSON response.
func (r *Request) JSON(code int, obj interface{}) error {
	r.Response.Header().Set("Content-Type", "application/json")
	r.StatusCode(code)
	return json.NewEncoder(r.Response).Encode(obj)
}

// GetMethod provides a simple way to return the request method type as a string
func (r *Request) GetMethod() string {
	return r.Method
}

// StatusCode sets the status code header
func (r *Request) StatusCode(c int) {
	r.Response.WriteHeader(c)
}
