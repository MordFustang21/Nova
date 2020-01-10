// Package nova is an HTTP request multiplexer. It matches the URL of each incoming request against a list of registered patterns
// and calls the handler for the pattern that most closely matches the URL. As well as providing some nice logging and response features.
package nova

import (
	"net/http"
	"path"
	"strings"
)

// Server represents the router and all associated data
type Server struct {
	// radix tree for looking up routes
	paths      map[string]*Node
	middleWare []Middleware

	// error callback func
	errorFunc ErrorFunc

	// debug defines logging for requests
	debug bool
}

// RequestFunc is the callback used in all handler func
type RequestFunc func(req *Request) error

// ErrorFunc is the callback used for errors
type ErrorFunc func(req *Request, err error)

// Node holds a single route with accompanying children routes
type Node struct {
	route    *Route
	children map[string]*Node
}

// Middleware holds all middleware functions
type Middleware struct {
	middleFunc func(*Request, func())
}

// New returns new supernova router
func New() *Server {
	return &Server{
		paths: map[string]*Node{},
		// set a default empty error func so we don't have to
		// check if it's set to nil
		errorFunc: func(req *Request, err error) {},
	}
}

// EnableDebug toggles output for incoming requests
func (sn *Server) EnableDebug(debug bool) {
	if debug {
		sn.debug = true
	}
}

// ErrorFunc sets the callback for errors
func (sn *Server) ErrorFunc(f ErrorFunc) {
	// only set if the passed value isn't nil
	if f != nil {
		sn.errorFunc = f
	}
}

// handler is the main entry point into the router
func (sn *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := NewRequest(w, r)
	if sn.debug {
		defer getDebugMethod(request)
	}

	// Run Middleware
	finished := sn.runMiddleware(request)
	if !finished {
		return
	}

	// search the tree for the route that matches the path and method
	route := sn.climbTree(request.GetMethod(), cleanPath(request.URL.Path))

	// if no route is found return a 404
	if route == nil {
		http.NotFound(request.ResponseWriter, request.Request)
		return
	}

	// execute the found route and if there is an error returned execute the error func
	err := route.call(request)
	if err != nil {
		sn.errorFunc(request, err)
	}
}

// All adds route for all http methods
func (sn *Server) All(route string, routeFunc RequestFunc) {
	sn.addRoute("", buildRoute(route, routeFunc))
}

// Get adds only GET method to route
func (sn *Server) Get(route string, routeFunc RequestFunc) {
	sn.addRoute(http.MethodGet, buildRoute(route, routeFunc))
}

// Post adds only POST method to route
func (sn *Server) Post(route string, routeFunc RequestFunc) {
	sn.addRoute(http.MethodPost, buildRoute(route, routeFunc))
}

// Put adds only PUT method to route
func (sn *Server) Put(route string, routeFunc RequestFunc) {
	sn.addRoute(http.MethodPut, buildRoute(route, routeFunc))
}

// Delete adds only DELETE method to route
func (sn *Server) Delete(route string, routeFunc RequestFunc) {
	sn.addRoute(http.MethodDelete, buildRoute(route, routeFunc))
}

// Restricted adds route that is restricted by method
func (sn *Server) Restricted(method, route string, routeFunc RequestFunc) {
	sn.addRoute(method, buildRoute(route, routeFunc))
}

// Group creates a new sub router that appends the path prefix
func (sn *Server) Group(path string) *RouteGroup {
	return &RouteGroup{
		s:    sn,
		path: path,
	}
}

// addRoute takes route and method and adds it to route tree
func (sn *Server) addRoute(method string, route *Route) {
	// if a base node isn't set for a method set it
	if sn.paths[method] == nil {
		sn.paths[method] = newNode()
	}

	// split the path parts and start building out tree from method node
	parts := strings.Split(route.route, "/")
	currentNode := sn.paths[method]
	for index, val := range parts {
		childKey := val
		if len(val) > 1 {
			// if first character is a colon this part of path is a parameter set to an empty key
			if val[0] == ':' {
				childKey = ""
			}
		}

		// see if node already exists
		if node, ok := currentNode.children[childKey]; ok {
			currentNode = node
		} else {
			n := newNode()
			currentNode.children[childKey] = n
			currentNode = n
		}

		// if at the last part of path set the child key to a new node
		// with the route set to the incoming route
		if index == len(parts)-1 {
			node := newNode()
			node.route = route
			currentNode.children[childKey] = node
		}
	}
}

func newNode() *Node {
	return &Node{
		children: map[string]*Node{},
	}
}

// climbTree takes in path and traverses tree to find route
func (sn *Server) climbTree(method, path string) *Route {
	parts := strings.Split(path, "/")

	currentNode, ok := sn.paths[method]
	if !ok {
		currentNode, ok = sn.paths[""]
		if !ok {
			return nil
		}
	}

	for _, val := range parts {
		var node *Node
		node = currentNode.children[val]
		if node == nil {
			node = currentNode.children[""]
		}

		if node == nil {
			return nil
		}

		currentNode = node
	}

	if node, ok := currentNode.children[parts[len(parts)-1]]; ok {
		return node.route
	}

	if node, ok := currentNode.children[""]; ok {
		return node.route
	}

	return nil
}

// buildRoute creates new Route
func buildRoute(route string, routeFunc RequestFunc) *Route {
	route = path.Clean(route)

	return &Route{
		routeFunc:        routeFunc,
		routeParamsIndex: map[int]string{},
		route:            route,
	}
}

// Use adds a new function to the middleware stack
func (sn *Server) Use(f func(req *Request, next func())) {
	if sn.middleWare == nil {
		sn.middleWare = make([]Middleware, 0)
	}

	sn.middleWare = append(sn.middleWare, Middleware{middleFunc: f})
}

// Internal method that runs the middleware
func (sn *Server) runMiddleware(req *Request) bool {
	stackFinished := true
	for m := range sn.middleWare {
		nextCalled := false
		sn.middleWare[m].middleFunc(req, func() {
			nextCalled = true
		})

		if !nextCalled {
			stackFinished = false
			break
		}
	}

	return stackFinished
}

// cleanPath returns the canonical path for p, eliminating . and .. elements.
// Borrowed from the net/http package.
func cleanPath(p string) string {
	if p == "" || p == "/" {
		return "/"
	}

	if p[0] != '/' {
		p = "/" + p
	}

	if p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}

	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}

	return np
}
