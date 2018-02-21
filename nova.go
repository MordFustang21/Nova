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
	s := new(Server)
	s.paths = make(map[string]*Node)
	return s
}

// EnableDebug toggles output for incoming requests
func (sn *Server) EnableDebug(debug bool) {
	if debug {
		sn.debug = true
	}
}

// Error sets the callback for errors
func (sn *Server) Error(f ErrorFunc) {
	sn.errorFunc = f
}

// handler is the main entry point into the router
func (sn *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := NewRequest(w, r)
	var logMethod func()
	if sn.debug {
		logMethod = getDebugMethod(request)
	}

	if logMethod != nil {
		defer logMethod()
	}

	// Run Middleware
	finished := sn.runMiddleware(request)
	if !finished {
		return
	}

	route := sn.climbTree(request.GetMethod(), cleanPath(request.URL.Path))
	if route != nil {
		err := route.call(request)
		// if we got error and erroFunc is set pass along
		if err != nil {
			if sn.errorFunc != nil {
				sn.errorFunc(request, err)
			}
		}

		return
	}

	http.NotFound(request.ResponseWriter, request.Request)
}

// All adds route for all http methods
func (sn *Server) All(route string, routeFunc RequestFunc) {
	sn.addRoute("", buildRoute(route, routeFunc))
}

// Get adds only GET method to route
func (sn *Server) Get(route string, routeFunc RequestFunc) {
	sn.addRoute("GET", buildRoute(route, routeFunc))
}

// Post adds only POST method to route
func (sn *Server) Post(route string, routeFunc RequestFunc) {
	sn.addRoute("POST", buildRoute(route, routeFunc))
}

// Put adds only PUT method to route
func (sn *Server) Put(route string, routeFunc RequestFunc) {
	sn.addRoute("PUT", buildRoute(route, routeFunc))
}

// Delete adds only DELETE method to route
func (sn *Server) Delete(route string, routeFunc RequestFunc) {
	sn.addRoute("DELETE", buildRoute(route, routeFunc))
}

// Restricted adds route that is restricted by method
func (sn *Server) Restricted(method, route string, routeFunc RequestFunc) {
	sn.addRoute(method, buildRoute(route, routeFunc))
}

// addRoute takes route and method and adds it to route tree
func (sn *Server) addRoute(method string, route *Route) {
	if sn.paths[method] == nil {
		sn.paths[method] = newNode()
	}

	parts := strings.Split(route.route, "/")
	currentNode := sn.paths[method]
	for index, val := range parts {
		childKey := val
		if len(val) > 1 {
			if val[0] == ':' {
				childKey = ""
			}
		}

		// see if node already exists
		if node, ok := currentNode.children[childKey]; ok {
			currentNode = node
		} else {
			node := newNode()
			currentNode.children[childKey] = node
			currentNode = node
		}

		if index == len(parts)-1 {
			node := newNode()
			node.route = route
			currentNode.children[childKey] = node
		}
	}
}

func newNode() *Node {
	n := new(Node)
	n.children = make(map[string]*Node)

	return n
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
	routeObj := new(Route)
	routeObj.routeFunc = routeFunc
	routeObj.routeParamsIndex = make(map[int]string)
	if route[len(route)-1] == '/' && len(route) > 1 {
		route = route[:len(route)-1]
	}

	routeObj.route = route

	return routeObj
}

// Use adds a new function to the middleware stack
func (sn *Server) Use(f func(req *Request, next func())) {
	if sn.middleWare == nil {
		sn.middleWare = make([]Middleware, 0)
	}
	middle := new(Middleware)
	middle.middleFunc = f
	sn.middleWare = append(sn.middleWare, *middle)
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
