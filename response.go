package nova

import (
	"net/http"
)

// Response is used to wrap http.ResponseWriter to collect response status code
type Response struct {
	http.ResponseWriter
	// status code
	Code int
}

// WriteHeader sets the status code and calls WriteHeader
func (sn *Response) WriteHeader(c int) {
	sn.Code = c
	sn.ResponseWriter.WriteHeader(c)
}
