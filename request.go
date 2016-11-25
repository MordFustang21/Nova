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
}

func NewRequest(r *http.Request) *Request {
	request := new(Request)
	request.R = r

	return request
}

func (r *Request) GetBody() []byte {
	data, err := ioutil.ReadAll(r.R.Body)
	if err != nil {
		println(err.Error())
	} else {
		return data
	}

	return make([]byte, 0)
}

func (r *Request) AsJson(i interface{}) error {

	jsn := r.GetBody()
	if len(jsn) != 0 {
		json.Unmarshal(jsn, i)
	} else {
		return errors.New("Request body is empty")
	}
	return nil
}

