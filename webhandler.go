package nova

import (
	"strings"
	"net/http"
)

type WebHandler struct {
	port       string
	Paths      []Route
	middleWare []MiddleWare
}

type MiddleWare struct {
	middleFunc func(*Request, *Response, Next)
}

type Next func()

//Returns new instance of the web handler
func Nova() *WebHandler {
	return new(WebHandler)
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//Build request response for middle ware
		request := NewRequest(r)
		response := new(Response)
		response.R = w

		//Run Middlware
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

		http.NotFound(w, r)

	})
	return http.ListenAndServe(":" + port, nil)
}

//Internal method that runs the middleware
func (wh *WebHandler) runMiddleware(request *Request, response *Response) {
	for m := range wh.middleWare {
		wh.middleWare[m].middleFunc(request, response, func() {
			return
		})
	}
}
