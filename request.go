package goExpress

import "net/http"

type Request struct {
	R     *http.Request
	RouteParams map[string]string
}

func NewRequest(r *http.Request) *Request {
	request := new(Request)
	request.R = r
	return request
}

