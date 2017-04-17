# Nova
A custom router for http.Server

Provides a lot of the same methods and functionality as Expressjs

Example
```go
package main

import (
	"github.com/MordFustang21/Nova"
)

func main() {
	// Get new instance of server
	s := supernova.New()

	//Static folder example
	s.AddStatic("/sitedir/")

	//If you want to cache a file (seconds)
	s.SetCacheTimeout(5)

	//Middleware Example
	s.Use(func(req *nova.Request, next func()) {
		req.Response.Header().Set("Powered-By", "supernova")
		next()
	})

	//Route Examples
	s.Post("/test/taco/:apple", func(req *nova.Request) {
		type test struct {
			Apple string
		}

		// Read JSON into struct from body
		var testS testS
		err := req.ReadJSON(&testS)
		if err != nil {
			log.Println(err)
		}
		req.Send("Received data")
	});

	// Example Get route with route params
	s.Get("/test/:taco/:apple", func(req *nova.Request) {
		tacoType := req.Param("taco")
		req.Send(tacoType)
	});

	// Resticted routes are used to restrict methods other than GET,PUT,POST,DELETE
	s.Restricted("OPTIONS", "/test/stuff", func(req *supernova.Request) {
		req.Send("OPTIONS Request received")
	})

	// Example post returning error
	s.Post("/register", func(req *nova.Request) {
		if len(req.Request.Body()) == 0 {
			// response code, error message, and any struct you want put into the errors array
			req.Error(500, "Body is empty", interface{})
		}
	})

	err := s.ListenAndServe(":8080")
	if err != nil {
		println(err.Error())
	}
}
```