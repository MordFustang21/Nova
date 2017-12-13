package nova

// Route is the construct of a single route pattern
type Route struct {
	routeFunc        RequestFunc
	routeParamsIndex map[int]string
	route            string
}

// call builds the route params & executes the function tied to the route
func (r *Route) call(req *Request) error {
	req.buildRouteParams(r.route)
	return r.routeFunc(req)
}
