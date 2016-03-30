package goExpress

import (
	"net/http"
	"encoding/json"
	"log"
)

type Response struct {
	response http.ResponseWriter
}

func (r *Response) Send(data interface{}) {
	switch v := data.(type) {
	case []byte:
		r.response.Write(v)
		break;
	case string:
		r.response.Write([]byte(v))
		break;
	}
}

func (r *Response) Json(obj interface{}) error {
	json, err := json.Marshal(obj)
	if err != nil {
		log.Println(err)
		return err
	} else {
		r.response.Header().Set("Content-Type", "application/json")
		r.response.Write(json)
	}
	return nil
}
