package nova

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

// Request resembles an incoming request
type Request struct {
	*http.Request
	Response    *Response
	routeParams map[string]string
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

	return req
}

// Param checks for and returns param or "" if doesn't exist
func (r *Request) Param(key string) string {
	if val, ok := r.routeParams[key]; ok {
		return val
	}

	return ""
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
	reqParts := strings.Split(r.RequestURI[1:], "/")
	routeParts := strings.Split(route[1:], "/")

	for index, val := range routeParts {
		if val[0] == ':' {
			r.routeParams[val[1:]] = reqParts[index]
		}
	}
}

// ReadJSON unmarshals request body into the struct provided
func (r *Request) ReadJSON(i interface{}) error {
	encoder := json.NewEncoder(r)
	return encoder.Encode(i)
}

// Send writes the data to the response body
func (r *Request) Send(data interface{}) (int, error) {
	switch v := data.(type) {
	case []byte:
		return r.Write(v)
	case string:
		return r.Write([]byte(v))
	case error:
		return r.Write([]byte(v.Error()))
	}

	return 0, errors.New("unsupported type")
}

// Write implements the Writer interface for http.Request.Body
func (r *Request) Write(b []byte) (int, error) {
	return r.Response.Write(b)
}

// Read implements reader for http.Request
func (r *Request) Read(p []byte) (int, error) {
	return r.Body.Read(p)
}

// Close implements Closer interface for http.Request.Body
func (r *Request) Close() error {
	return r.Request.Body.Close()
}

// JSON marshals the given interface object and writes the JSON response.
func (r *Request) JSON(code int, obj interface{}) error {
	encoder := json.NewEncoder(r)

	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.WriteHeader(code)
	return encoder.Encode(obj)
}

// GetMethod provides a simple way to return the request method type as a string
func (r *Request) GetMethod() string {
	return r.Method
}

// buildUrlParams builds url params and returns base route
func (r *Request) buildUrlParams() {
	reqUrl := string(r.RequestURI)
	baseParts := strings.Split(reqUrl, "?")

	if len(baseParts) == 0 {
		return
	}

	params := strings.Join(baseParts[1:], "")
	paramParts := strings.Split(params, "&")

	for i := range paramParts {
		keyValue := strings.Split(paramParts[i], "=")
		if len(keyValue) > 1 {
			r.routeParams[keyValue[0]] = keyValue[1]
		}
	}
}

// GetPusher returns http.Pusher interface if exists or nil on HTTP/1.1 connections
func (r *Request) GetPusher() http.Pusher {
	if pusher, ok := r.Response.ResponseWriter.(http.Pusher); ok {
		return pusher
	}

	return nil
}
