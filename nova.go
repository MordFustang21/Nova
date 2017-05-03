// Package nova is a http.Server router that implements a radix tree for fast lookups
// and adds middleware and other helpful features to build APIs
package nova

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Server represents the router and all associated data
type Server struct {
	server *http.Server
	ln     net.Listener

	// radix tree for looking up routes
	paths map[string]*Node

	staticDirs         []staticAsset
	middleWare         []Middleware
	cachedStatic       *CachedStatic
	maxCachedTime      int64
	compressionEnabled bool

	// debug defines logging for requests
	debug bool
}

// Node holds a single route with accompanying children routes
type Node struct {
	route    *Route
	children map[string]*Node
}

// CachedObj represents a static asset
type CachedObj struct {
	data       []byte
	timeCached time.Time
}

// CachedStatic holds all cached static assets in memory
type CachedStatic struct {
	mutex sync.Mutex
	files map[string]*CachedObj
}

// staticAsset contains all info for static dir
type staticAsset struct {
	// directory of static files
	path string

	// used for http2 push
	pushAssets []string
}

// Middleware holds all middleware functions
type Middleware struct {
	middleFunc func(*Request, func())
}

// New returns new supernova router
func New() *Server {
	s := new(Server)
	s.server = &http.Server{
		Handler: s,
	}
	s.cachedStatic = new(CachedStatic)
	s.cachedStatic.files = make(map[string]*CachedObj)
	return s
}

// EnableDebug toggles output for incoming requests
func (sn *Server) EnableDebug(debug bool) {
	if debug {
		sn.debug = true
	}
}

// ListenAndServe starts the server
func (sn *Server) ListenAndServe(addr string) error {
	sn.server.Addr = addr
	return sn.server.ListenAndServe()
}

// ServeTLS starts server with ssl
func (sn *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
	sn.server.Addr = addr
	return sn.server.ListenAndServeTLS(certFile, keyFile)
}

// Close closes existing listener
func (sn *Server) Close() error {
	return sn.ln.Close()
}

// handler is the main entry point into the router
func (sn *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := NewRequest(w, r)
	var logMethod func()
	if sn.debug {
		logMethod = getDebugMethod(request)
		defer logMethod()
	}

	// Run Middleware
	finished := sn.runMiddleware(request)
	if !finished {
		return
	}

	node := sn.climbTree(request.GetMethod(), request.RequestURI)
	if node != nil {
		node.route.call(request)
		return
	}

	// Check for static file
	served := sn.serveStatic(request)
	if served {
		return
	}

	http.NotFound(request.Response, request.Request)
}

// All adds route for all http methods
func (sn *Server) All(route string, routeFunc func(*Request)) {
	sn.addRoute("", buildRoute(route, routeFunc))
}

// Get adds only GET method to route
func (sn *Server) Get(route string, routeFunc func(*Request)) {
	sn.addRoute("GET", buildRoute(route, routeFunc))
}

// Post adds only POST method to route
func (sn *Server) Post(route string, routeFunc func(*Request)) {
	sn.addRoute("POST", buildRoute(route, routeFunc))
}

// Put adds only PUT method to route
func (sn *Server) Put(route string, routeFunc func(*Request)) {
	sn.addRoute("PUT", buildRoute(route, routeFunc))
}

// Delete adds only DELETE method to route
func (sn *Server) Delete(route string, routeFunc func(*Request)) {
	sn.addRoute("DELETE", buildRoute(route, routeFunc))
}

// Restricted adds route that is restricted by method
func (sn *Server) Restricted(method, route string, routeFunc func(*Request)) {
	sn.addRoute(method, buildRoute(route, routeFunc))
}

// addRoute takes route and method and adds it to route tree
func (sn *Server) addRoute(method string, route *Route) {
	routeStr := route.route
	if routeStr[len(routeStr)-1] == '/' {
		routeStr = routeStr[:len(routeStr)-1]
		route.route = routeStr
	}
	if sn.paths == nil {
		sn.paths = make(map[string]*Node)
	}

	if sn.paths[method] == nil {
		node := new(Node)
		node.children = make(map[string]*Node)
		sn.paths[method] = node
	}

	parts := strings.Split(routeStr[1:], "/")

	currentNode := sn.paths[method]
	for index, val := range parts {
		childKey := val
		if val[0] == ':' {
			childKey = ""
		} else {
			childKey = val
		}

		if node, ok := currentNode.children[childKey]; ok {
			currentNode = node
		} else {
			node := getNode()
			currentNode.children[childKey] = node
			currentNode = node
		}

		if index == len(parts)-1 {
			currentNode.route = route
		}
	}
}

// getNode builds a new node to be added to the radix tree
func getNode() *Node {
	node := new(Node)
	node.children = make(map[string]*Node)
	return node
}

// climbTree takes in path and traverses tree to find route
func (sn *Server) climbTree(method, path string) *Node {
	parts := strings.Split(path[1:], "/")
	pathLen := len(parts)

	var currentNode *Node
	var ok bool
	if currentNode, ok = sn.paths[method]; !ok {
		currentNode, ok = sn.paths[""]
		if !ok {
			return nil
		}
	}

	for i := 0; i < pathLen; i++ {
		node, found := currentNode.children[parts[i]]
		if !found {
			node = currentNode.children[""]
		}

		// path not found return to avoid wasting time on long routes
		if node == nil {
			return nil
		}

		currentNode = node
	}

	return currentNode
}

// buildRoute creates new *Route
func buildRoute(route string, routeFunc func(*Request)) *Route {
	routeObj := new(Route)
	routeObj.routeFunc = routeFunc
	routeObj.routeParamsIndex = make(map[int]string)
	routeObj.route = route

	return routeObj
}

// AddStatic adds static route to be served
func (sn *Server) AddStatic(dir string, pushableAssets ...string) {
	if _, err := os.Stat(dir); err == nil {
		asset := staticAsset{path: dir, pushAssets: pushableAssets}
		sn.staticDirs = append(sn.staticDirs, asset)
	}
}

// EnableGzip turns on Gzip compression for static
func (sn *Server) EnableGzip(value bool) {
	sn.compressionEnabled = value
}

// serveStatic looks up folder and serves static files
func (sn *Server) serveStatic(req *Request) bool {
	for _, staticDir := range sn.staticDirs {
		path := staticDir.path + req.URL.Path

		// Sanitize path
		path = strings.Replace(path, "..", "", -1)

		// If ends in / default to index.html

		if path[len(path)-1] == '/' {
			path += "index.html"
		}

		stat, err := os.Stat(path)
		if err != nil {
			continue
		}

		// Set mime type
		extensionParts := strings.Split(path, ".")
		ext := extensionParts[len(extensionParts)-1]
		mType := mime.TypeByExtension("." + ext)

		if mType != "" {
			req.Response.Header().Set("Content-Type", mType)
		}

		if sn.compressionEnabled && stat.Size() < 10000000 {
			var b bytes.Buffer
			writer := gzip.NewWriter(&b)

			data, err := ioutil.ReadFile(path)
			if err != nil {
				println("Unable to read: " + err.Error())
			}

			writer.Write(data)
			writer.Close()
			req.Response.Header().Set("Content-Encoding", "gzip")
			req.Send(b.String())
		} else {
			file, err := os.Open(path)
			if err != nil {
				return false
			}

			// Try to push http2 assets associated with folder
			if p := req.GetPusher(); p != nil {
				for _, val := range staticDir.pushAssets {
					p.Push(val, nil)
				}
			}

			io.Copy(req, file)
		}

		return true
	}
	return false
}

// Use adds a new function to the middleware stack
func (sn *Server) Use(f func(*Request, func())) {
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

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shut down.
// If the provided context expires before the shutdown is complete,
// then the context's error is returned.
//
// Shutdown does not attempt to close nor wait for hijacked
// connections such as WebSockets. The caller of Shutdown should
// separately notify such long-lived connections of shutdown and wait
// for them to close, if desired.
func (sn *Server) Shutdown(ctx context.Context) error {
	return sn.server.Shutdown(ctx)
}
