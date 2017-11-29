package nova

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"fmt"
	"golang.org/x/net/context"
)

// Request resembles an incoming request
type Request struct {
	*http.Request
	ResponseWriter http.ResponseWriter
	routeParams    map[string]string
	urlValues      url.Values
	BaseUrl        string
	Ctx            context.Context
	ResponseCode   int
}

// JSONError resembles the RESTful standard for an error response
type JSONError struct {
	Errors  []string `json:"errors"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
}

// JSONErrors holds the JSONError response
type JSONErrors struct {
	Error JSONError `json:"error"`
}

// NewRequest creates a new Request pointer for an incoming request
func NewRequest(w http.ResponseWriter, r *http.Request) *Request {
	req := new(Request)
	req.Request = r
	req.routeParams = make(map[string]string)
	req.ResponseWriter = w
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

// Error provides and easy way to send a structured error response
// Error will use error and fmt.Stringer interface otherwise fmt %v
func (r *Request) Error(statusCode int, msg string, errs ...interface{}) error {
	var errList []string

	// Attempt to get string of errors
	for _, err := range errs {
		switch v := err.(type) {
		case error:
			errList = append(errList, v.Error())
		case fmt.Stringer:
			errList = append(errList, v.String())
		default:
			errList = append(errList, fmt.Sprintf("%v", v))
		}
	}

	// Format error response
	newErr := JSONErrors{
		Error: JSONError{
			Errors:  errList,
			Code:    statusCode,
			Message: msg,
		},
	}

	return r.JSON(statusCode, newErr)
}

// buildRouteParams builds a map of the route params
func (r *Request) buildRouteParams(route string) {
	routeParams := r.routeParams
	reqParts := strings.Split(r.BaseUrl, "/")
	routeParts := strings.Split(route, "/")

	for index, val := range routeParts {
		if len(val) > 1 {
			if val[0] == ':' {
				routeParams[val[1:]] = reqParts[index]
			}
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
		_, err = r.ResponseWriter.Write(v)
	case string:
		_, err = r.ResponseWriter.Write([]byte(v))
	case error:
		_, err = r.ResponseWriter.Write([]byte(v.Error()))
	default:
		err = errors.New("unsupported type Send type")
	}

	return err
}

// JSON marshals the given interface object and writes the JSON response.
func (r *Request) JSON(code int, obj interface{}) error {
	r.ResponseWriter.Header().Set("Content-Type", "application/json")
	r.StatusCode(code)
	return json.NewEncoder(r.ResponseWriter).Encode(obj)
}

// GetMethod provides a simple way to return the request method type as a string
func (r *Request) GetMethod() string {
	return r.Method
}

// StatusCode sets the status code header
func (r *Request) StatusCode(c int) {
	r.WriteHeader(c)
}

func (r *Request) WriteHeader(c int) {
	r.ResponseWriter.WriteHeader(c)
}

func (r *Request) Header() http.Header {
	return r.ResponseWriter.Header()
}
