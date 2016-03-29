package goExpress

import (
	"net/http"
	"encoding/json"
	"log"
)

type Response struct {
	response http.ResponseWriter
}

func (r *Response) Send(msg string) {
	r.response.Write([]byte(msg))
}

func (r *Response) Json(obj interface{}) {
	json, err := json.Marshal(obj)
	if err != nil {
		log.Println(err)
	} else {
		r.response.Write(json)
	}

}
