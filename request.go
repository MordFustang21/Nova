package nova

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// Request resembles an incoming request
type Request struct {
	*http.Request
	ResponseWriter http.ResponseWriter
	routeParams    map[string]string
	queryParams    url.Values
	BaseUrl        string
	ResponseCode   int
}

// JSONError resembles the RESTful standard for an error response
type JSONError struct {
	Code    int      `json:"code"`
	Errors  []string `json:"errors"`
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
	req.queryParams = r.URL.Query()
	req.BaseUrl = r.RequestURI

	req.ResponseCode = http.StatusOK

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
	return r.queryParams.Get(key)
}

// Error provides and easy way to send a structured error response
func (r *Request) Error(statusCode int, msg string, userErr error) error {
	// Format error response
	newErr := JSONErrors{
		Error: JSONError{
			Code:    statusCode,
			Message: msg,
		},
	}

	// json encode the response and if there is an error encoding wrap the users error
	// and return it
	err := r.JSON(statusCode, newErr)
	if err != nil {
		if userErr == nil {
			return errors.Wrap(err, "unable to marshal the response")
		}

		return errors.Wrap(userErr, err.Error())
	}

	return userErr
}

// buildRouteParams builds a map of the route params
func (r *Request) buildRouteParams(route string) {
	routeParams := r.routeParams
	reqParts := strings.Split(r.BaseUrl, "/")
	routeParts := strings.Split(route, "/")

	for index, val := range routeParts {
		if len(val) > 1 {
			if val[0] == ':' {
				param := strings.Split(reqParts[index], "?")
				routeParams[val[1:]] = param[0]
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

// Write will write data to the response and set a given status code
func (r *Request) Write(code int, data interface{}) error {
	// set status code for response
	r.StatusCode(code)

	return r.Send(data)
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
	r.ResponseCode = c
	r.WriteHeader(c)
}

// WriteHeader sends an HTTP response header with the provided
// status code.
//
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes.
//
// The provided code must be a valid HTTP 1xx-5xx status code.
// Only one header may be written. Go does not currently
// support sending user-defined 1xx informational headers,
// with the exception of 100-continue response header that the
// Server sends automatically when the Request.Body is read.
func (r *Request) WriteHeader(c int) {
	r.ResponseWriter.WriteHeader(c)
}

// Header returns the header map that will be sent by
// WriteHeader. The Header map also is the mechanism with which
// Handlers can set HTTP trailers.
//
// Changing the header map after a call to WriteHeader (or
// Write) has no effect unless the modified headers are
// trailers.
//
// There are two ways to set Trailers. The preferred way is to
// predeclare in the headers which trailers you will later
// send by setting the "Trailer" header to the names of the
// trailer keys which will come later. In this case, those
// keys of the Header map are treated as if they were
// trailers. See the example. The second way, for trailer
// keys not known to the Handler until after the first Write,
// is to prefix the Header map keys with the TrailerPrefix
// constant value. See TrailerPrefix.
//
// To suppress automatic response headers (such as "Date"), set
// their value to nil.
func (r *Request) Header() http.Header {
	return r.ResponseWriter.Header()
}
