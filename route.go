package goExpress

import "strings"

type Route struct {
	rq               *Request
	rs               *Response
	rr               RequestResponse
	routeParamsIndex map[int]string
}

func (r *Route) buildRouteParams() {
	routeParams := make(map[string]string)
	pathParts := strings.Split(r.rq.request.URL.Path, "/")

	for i := range r.routeParamsIndex {
		name := r.routeParamsIndex[i]
		if i <= len(pathParts) - 1 {
			routeParams[name] = pathParts[i]
		}
	}

	r.rq.routeParams = routeParams
}

func (r *Route) prepare() {
	r.buildRouteParams()
}

func (r *Route) call() {
	r.rr(r.rq, r.rs)
}
