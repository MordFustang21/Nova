package goExpress

import (
	"strings"
	"net/http"
	"fmt"
)

type WebHandler struct {
	port  string
	Paths []Route
}

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

	fmt.Println(baseDir)
	routeObj.route = baseDir

	if wh.Paths == nil {
		wh.Paths = make([]Route, 0)
	}

	wh.Paths = append(wh.Paths, *routeObj)
}

func (wh *WebHandler) Serve(port string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//Split and build first full path
		pathParts := strings.Split(r.URL.Path, "/")
		path := strings.Join(pathParts, "/")

		//Check all paths
		for _ = range pathParts {
			fmt.Println("Checking ", path)
			for routeIndex := range wh.Paths {
				route := wh.Paths[routeIndex]
				if route.route == path || route.route == path + "/" {
					request := NewRequest(r)
					response := new(Response)
					response.response = w

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
