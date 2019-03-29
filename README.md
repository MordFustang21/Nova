![Supernova Logo](https://raw.githubusercontent.com/MordFustang21/supernova-logo/master/nova_logo.png)

[![GoDoc](https://godoc.org/github.com/MordFustang21/nova?status.svg)](https://godoc.org/github.com/MordFustang21/nova)
[![Go Report Card](https://goreportcard.com/badge/github.com/mordfustang21/nova)](https://goreportcard.com/report/github.com/mordfustang21/nova)
[![Build Status](https://travis-ci.org/MordFustang21/nova.svg)](https://travis-ci.org/MordFustang21/nova)

nova is a mux for http while we don't claim to be the best or fastest we provide a lot of tools and features that enable
you to be highly productive and help build up your api quickly and efficiently.

*Note nova's exported API interface will continue to change in unpredictable, backwards-incompatible ways until we tag a v1.0.0 release.

### Start using it
1. Download and install
```
$ go get github.com/MordFustang21/nova
```
2. Import it into your code
```
import "github.com/MordFustang21/nova"
```

### Basic Usage
http://localhost:8080/hello
```go
package main

import (
	"log"
	"net/http"
	
	"github.com/MordFustang21/nova"
	)

func main() {
	s := nova.New()
	
	s.Get("/hello", func(request *nova.Request) error {
	    return request.Send("world")
	})
	
	if err := http.ListenAndServe(":8080", s); err != nil {
    		log.Fatal(err)
	}
}

```
### Route Group
This will create a route with a base path than you can append other paths onto.

This example creates two routes "/v1/hello" and "/v2/world" so you can keep backwards compatible changes
```go
package main

import (
	"log"
	"net/http"
	
	"github.com/MordFustang21/nova"
	)

func worldRequest(request *nova.Request) error {
	return request.Send("world")
}

func main() {
	s := nova.New()
	
	v1Group := s.Group("/v1")
	v2Group := s.Group("/v2")
	
	v1Group.Get("/hello", worldRequest)
	v2Group.Get("/world", worldRequest)
	
	if err := http.ListenAndServe(":8080", s); err != nil {
		log.Fatal(err)
	}
}

```
#### Retrieving parameters
http://localhost:8080/hello/world
```go
package main

import (
	"log"
	"net/http"
	
	"github.com/MordFustang21/nova"
	)

func main() {
	s := nova.New()
	
	s.Get("/hello/:text", func(request *nova.Request) error {
		t := request.RouteParam("text")
	    return request.Send(t)
	})
	
	if err := http.ListenAndServe(":8080", s); err != nil {
    		log.Fatal(err)
	}
}
```

#### Returning Errors
http://localhost:8080/hello
```go
package main

import (
	"log"
	"net/http"
	
	"github.com/MordFustang21/nova"
	)

func main() {
	s := nova.New()
	
	s.Post("/hello", func(request *nova.Request) error {
		r := struct {
		 World string
		}{}
		
		// ReadJSON will attempt to unmarshall the json from the request body into the given struct
		err := request.ReadJSON(&r)
		if err != nil {
		    return request.Error(http.StatusBadRequest, "couldn't parse request", err.Error())
		}
		
		// JSON will marshall the given object and marshall into into the response body
		return request.JSON(http.StatusOK, r)
	})
	
	if err := http.ListenAndServe(":8080", s); err != nil {
    		log.Fatal(err)
	}
	
}
```