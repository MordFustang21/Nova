package nova

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"errors"
)

type Request struct {
	R           *http.Request
	RouteParams map[string]string
	Body        []byte
}

func NewRequest(r *http.Request) *Request {
	request := new(Request)
	request.R = r
	if r.Body != nil {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {

		} else {
			request.Body = data
		}
	}

	return request
}

func (r *Request) Json(i interface{}) error {
	if r.Body == nil {
		return errors.New("Request Body is empty")
	} else {
		return json.Unmarshal(r.Body, i)
	}
}

