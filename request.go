package goExpress

import "net/http"

type Request struct {
	request     *http.Request
	routeParams map[string]string
}

func NewRequest(r *http.Request) *Request {
	request := new(Request)
	request.request = r
	return request
}

