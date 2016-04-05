package nova

import (
	"strings"
	"net/http"
	"os"
	"io/ioutil"
	"log"
	"mime"
)

type WebHandler struct {
	port       string
	Paths      []Route
	middleWare []MiddleWare
	staticDirs []string
}

type MiddleWare struct {
	middleFunc func(*Request, *Response, Next)
}

type Next func()

//Returns new instance of the web handler
func Nova() *WebHandler {
	wh := new(WebHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//Build request response for middle ware
		request := NewRequest(r)
		response := new(Response)
		response.R = w

		//Run Middleware
		wh.runMiddleware(request, response)

		//Split and build first full path
		pathParts := strings.Split(r.URL.Path, "/")
		path := strings.Join(pathParts, "/")

		//Check all paths
		for _ = range pathParts {
			for routeIndex := range wh.Paths {
				route := wh.Paths[routeIndex]
				if route.route == path || route.route == path + "/" {
					route.rq = request
					route.rs = response

					//Prepare data for call
					route.prepare()

					//Call user handler
					route.call()
					return
				}
			}

			//Cut end of array
			_, pathParts = pathParts[len(pathParts) - 1], pathParts[:len(pathParts) - 1]
			path = strings.Join(pathParts, "/")
		}

		//Check for static file
		served := wh.serveStatic(request, response)

		if served {
			return
		}

		http.NotFound(w, r)

	})

	return wh
}

//Adds new route to the stack
func (wh *WebHandler) AddRoute(route string, rr RequestResponse) {
	routeObj := new(Route)
	routeObj.rr = rr

	routeObj.routeParamsIndex = make(map[int]string)

	routeParts := strings.Split(route, "/")
	baseDir := ""
	for i := range routeParts {
		if strings.Contains(routeParts[i], ":") {
			routeParamMod := strings.Replace(routeParts[i], ":", "", 1)
			routeObj.routeParamsIndex[i] = routeParamMod
		} else {
			baseDir += routeParts[i] + "/"
		}

	}

	routeObj.route = baseDir

	if wh.Paths == nil {
		wh.Paths = make([]Route, 0)
	}

	wh.Paths = append(wh.Paths, *routeObj)
}

//Used to serve static files
func (wh *WebHandler) AddStatic(dir string) {

	if wh.staticDirs == nil {
		wh.staticDirs = make([]string, 0)
	}

	if _, err := os.Stat(dir); err == nil {
		wh.staticDirs = append(wh.staticDirs, dir)
	}
}

//Adds a new function to the middleware stack
func (wh *WebHandler) Use(f func(*Request, *Response, Next)) {
	if wh.middleWare == nil {
		wh.middleWare = make([]MiddleWare, 0)
	}
	middle := new(MiddleWare)
	middle.middleFunc = f
	wh.middleWare = append(wh.middleWare, *middle)
}

//Starts serving the routes
func (wh *WebHandler) Serve(port string) error {
	return http.ListenAndServe(":" + port, nil)
}

//Start serving secure
func (wh *WebHandler) ServeSecure(addr string, certFile string, keyFile string) error {
	return http.ListenAndServeTLS(":" + addr, certFile, keyFile, nil)
}

//Internal method that runs the middleware
func (wh *WebHandler) runMiddleware(request *Request, response *Response) {
	for m := range wh.middleWare {
		wh.middleWare[m].middleFunc(request, response, func() {
			return
		})
	}
}

//TODO: Convert to middleware
func (wh *WebHandler) serveStatic(req *Request, res *Response) bool {

	for i := range wh.staticDirs {
		staticDir := wh.staticDirs[i]
		path := staticDir + req.R.URL.Path

		//Remove all .. for security TODO: Allow if doesn't go above basedir
		path = strings.Replace(path, "..", "", -1)

		//If ends in / default to index.html
		if strings.HasSuffix(path, "/") {
			path += "index.html"
		}

		if _, err := os.Stat(path); err == nil {

			contents, err := ioutil.ReadFile(path)

			if err != nil {
				log.Println("Unable to serve file")
			}

			//Set mime type
			extensionParts := strings.Split(path, ".")
			ext := extensionParts[len(extensionParts) - 1]
			mType := mime.TypeByExtension("." + ext)

			if mType != "" {
				res.R.Header().Set("Content-Type", mType)
			}

			res.Send(contents)
			return true
		}
	}

	return false
}