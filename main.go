package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/xid"

	"github.com/shelakel/go-exp-mvc/web"
	"golang.org/x/net/context"
)

func main() {
	connect := web.ComposeMiddleware(logRequest)

	// call connect before listen to run runtime checks for compatible middleware/handlers
	defaultHandler := connect(helloWorld)
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.Header.Get("X-Request-ID")) == 0 {
			r.Header.Set("X-Request-ID", xid.New().String())
		}
		c := &web.Context{
			Context:  context.Background(),
			Request:  r,
			Response: w,
		}
		defaultHandler.ServeHTTP(c)
	}))
}

func logRequest(next web.Handler) web.Handler {
	return web.HandlerFunc(func(c *web.Context) {
		r := c.Request
		defer func(start time.Time) {
			elapsed := time.Since(start)
			fmt.Printf("%s %s (%.2f ms)\n", r.Method, r.RequestURI, float64(elapsed.Nanoseconds())/1000000)
		}(time.Now())
		next.ServeHTTP(c)
	})
}

func helloWorld(c context.Context, w http.ResponseWriter, r *http.Request) web.Handler {
	return web.Result("Hello world", 200)
}
