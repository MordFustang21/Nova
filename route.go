package nova

import (
	"net/http"
	"path"
)

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

// RouteGroup is used to add routes prepending a base path
type RouteGroup struct {
	// server to add the route to
	s *Server

	// base path to prepend the path
	path string
}

// All adds route for all http methods
func (r *RouteGroup) All(route string, routeFunc RequestFunc) {
	route = path.Join(r.path, route)
	r.s.addRoute("", buildRoute(route, routeFunc))
}

// Get adds only GET method to route
func (r *RouteGroup) Get(route string, routeFunc RequestFunc) {
	route = path.Join(r.path, route)
	r.s.addRoute(http.MethodGet, buildRoute(route, routeFunc))
}

// Post adds only POST method to route
func (r *RouteGroup) Post(route string, routeFunc RequestFunc) {
	route = path.Join(r.path, route)
	r.s.addRoute(http.MethodPost, buildRoute(route, routeFunc))
}

// Put adds only PUT method to route
func (r *RouteGroup) Put(route string, routeFunc RequestFunc) {
	route = path.Join(r.path, route)
	r.s.addRoute(http.MethodPut, buildRoute(route, routeFunc))
}

// Delete adds only DELETE method to route
func (r *RouteGroup) Delete(route string, routeFunc RequestFunc) {
	route = path.Join(r.path, route)
	r.s.addRoute(http.MethodDelete, buildRoute(route, routeFunc))
}

// Restricted adds route that is restricted by method
func (r *RouteGroup) Restricted(method, route string, routeFunc RequestFunc) {
	route = path.Join(r.path, route)
	r.s.addRoute(method, buildRoute(route, routeFunc))
}
