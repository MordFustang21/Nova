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
	Response    http.ResponseWriter
	routeParams map[string]string
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
	req.Response = w
	req.routeParams = make(map[string]string)
	req.BaseUrl = r.RequestURI

	return req
}

// Param checks for and returns param or "" if doesn't exist
func (r *Request) Param(key string) string {
	if val, ok := r.routeParams["string"]; ok {
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
	encoder := json.NewEncoder(r.Response)
	return encoder.Encode(i)
}

// Send writes the data to the response body
func (r *Request) Send(data interface{}) (int, error) {
	switch v := data.(type) {
	case []byte:
		return r.Response.Write(v)
	case string:
		return r.Response.Write([]byte(v))
	case error:
		return r.Response.Write([]byte(v.Error()))
	}

	return 0, errors.New("unsupported type")
}

// JSON marshals the given interface object and writes the JSON response.
func (r *Request) JSON(code int, obj interface{}) error {
	encoder := json.NewEncoder(r.Response)

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
