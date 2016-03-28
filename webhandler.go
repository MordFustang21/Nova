package goExpress

import (
	"strings"
	"net/http"
)

type WebHandler struct {
	port  string
	//TODO: Create as slice so lookups happen in order
	Paths map[string]Route
}

func (wh *WebHandler) AddRoute(route string, rr RequestResponse) {
	routeObj := new(Route)
	routeObj.rr = rr

	routeObj.routeParamsIndex = make(map[int]string)

	routeParts := strings.Split(route, "/")

	for i := range routeParts {
		if i > 1 {
			routeParamMod := strings.Replace(routeParts[i], ":", "", 1)
			routeObj.routeParamsIndex[i] = routeParamMod
		}
	}

	//TODO: Combine all parts that don't contain :
	baseDir := "/" + routeParts[1]
	wh.Paths[baseDir] = *routeObj
	wh.Paths[baseDir + "/"] = *routeObj
}

func (wh *WebHandler) Serve(port string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		path := "/"
		for i := range pathParts {
			if pathParts[i] != "" {
				path += pathParts[i] + "/"
			}
			route, exists := wh.Paths[path]
			if exists {
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

		http.NotFound(w, r)

	})
	return http.ListenAndServe(":" + port, nil)
}
