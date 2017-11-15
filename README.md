![nova Logo](https://raw.githubusercontent.com/MordFustang21/supernova-logo/master/supernova_banner.png)

[![GoDoc](https://godoc.org/github.com/MordFustang21/nova?status.svg)](https://godoc.org/github.com/MordFustang21/nova)
[![Go Report Card](https://goreportcard.com/badge/github.com/mordfustang21/nova)](https://goreportcard.com/report/github.com/mordfustang21/nova)
[![Build Status](https://travis-ci.org/MordFustang21/nova.svg)](https://travis-ci.org/MordFustang21/nova)

nova is a mux for http While we don't claim to be the best or fastest we are still very fast thanks to [fasthttp](https://github.com/valyala/fasthttp)
and help you be highly productive by providing easy to use tools that help build up your api quickly.

*Note nova's exported API interface will continue to change in unpredictable, backwards-incompatible ways until we tag a v2.0.0 release.

### Start using it
1. Download and install
```
$ go get github.com/MordFustang21/nova
```
2. Import it into your code
```
import "github.com/MordFustang21/nova"
```

### Use a vendor tool like dep
1. go get dep
```
$ go get -u github.com/golang/dep/cmd/dep
```
2. cd to project folder and run dep
```
$ dep ensure
```

Refer to [dep](https://github.com/golang/dep) for more information

### Basic Usage
http://localhost:8080/hello
```go
package main

import "github.com/MordFustang21/nova"

func main() {
	s := nova.New()
	
	s.Get("/hello", func(request *nova.Request) {
	    request.Send("world")
	    return
	})
	
	s.ListenAndServe(":8080")
}

```
#### Retrieving parameters
http://localhost:8080/hello/world
```go
package main

import "github.com/MordFustang21/nova"

func main() {
	s := nova.New()
	
	s.Get("/hello/:text", func(request *nova.Request) {
		t := request.RouteParam("text")
	    request.Send(t)
	})
	
	s.ListenAndServe(":8080")
}
```

#### Returning Errors
http://localhost:8080/hello
```go
package main

import (
	"net/http"
	"github.com/MordFustang21/nova"
)

func main() {
	s := nova.New()
	
	s.Post("/hello", func(request *nova.Request) {
		r := struct {
		 World string
		}{}
		
		// ReadJSON will attempt to unmarshall the json from the request body into the given struct
		err := request.ReadJSON(&r)
		if err != nil {
		    request.Error(http.StatusBadRequest, "couldn't parse request", err.Error())
		    return
		}
		
		// JSON will marshall the given object and marshall into into the response body
		request.JSON(http.StatusOK, r)
	})
	
	s.ListenAndServe(":8080")
	
}
```