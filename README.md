# go-express (name will change)
An express like framework for go web servers

Provides a lot of the same methods as express.js

Example
```go
wh := new(goExpress.WebHandler)

wh.AddRoute("/test/taco/:apple", func(req *goExpress.Request, res *goExpress.Response) {
	res.Send("Received Taco")
});

wh.AddRoute("/test/:taco/:apple", func(req *goExpress.Request, res *goExpress.Response) {
	res.Json(req.RouteParams)
});

err := wh.Serve("8080")

if err != nil {
	log.Fatal(err)
}
```
