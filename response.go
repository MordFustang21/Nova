package goExpress

import "net/http"

type Response struct {
	response http.ResponseWriter
}

func (r *Response) Send(msg string) {
	r.response.Write([]byte(msg))
}
