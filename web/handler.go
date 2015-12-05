package web

import (
	"fmt"
	"net/http"
	"reflect"

	"golang.org/x/net/context"
)

// Context represents a HTTP Context.
type Context struct {
	Context  context.Context
	Request  *http.Request
	Response http.ResponseWriter
}

// Handler is a http.Handler with a net/context.Context parameter.
type Handler interface {
	ServeHTTP(c *Context)
}

// HandlerFunc implements Handler.
type HandlerFunc func(c *Context)

func (f HandlerFunc) ServeHTTP(c *Context) {
	f(c)
}

// HandlerSpreadFunc implements Handler.
type HandlerSpreadFunc func(c context.Context, w http.ResponseWriter, r *http.Request)

func (f HandlerSpreadFunc) ServeHTTP(c *Context) {
	f(c.Context, c.Response, c.Request)
}

// ResultHandler is a Handler returning a Handler (action result)
type ResultHandler interface {
	ServeHTTP(c *Context) Handler
}

// ResultHandlerFunc implements ResultHandler.
type ResultHandlerFunc func(c *Context) Handler

func (f ResultHandlerFunc) ServeHTTP(c *Context) Handler {
	return f(c)
}

// ResultHandlerSpreadFunc implements ResultHandler.
type ResultHandlerSpreadFunc func(c context.Context, w http.ResponseWriter, r *http.Request) Handler

func (f ResultHandlerSpreadFunc) ServeHTTP(c *Context) Handler {
	return f(c.Context, c.Response, c.Request)
}

// ComposeMiddleware combines Middleware into a single Middleware function
func ComposeMiddleware(middleware ...interface{}) func(interface{}) Handler {
	if middleware == nil || len(middleware) == 0 {
		return convertHandlerPanic
	}
	// convert before returning to fail early at ComposeMiddleware call
	adapters := make([]func(Handler) Handler, len(middleware))
	for i := 0; i < len(middleware); i++ {
		if middleware[i] == nil {
			panic(fmt.Sprintf("Middleware of type '%v' at index %d is nil.", reflect.TypeOf(middleware[i]), i))
		}
		adapters[i] = convertMiddleware(middleware, i)
	}
	return func(v interface{}) Handler {
		next := convertHandlerPanic(v)
		for i := len(adapters) - 1; i >= 0; i-- {
			next = adapters[i](next)
		}
		return next
	}
}

// middleware

func convertMiddleware(middleware []interface{}, i int) func(Handler) Handler {
	switch mw := middleware[i].(type) {
	case func(interface{}) Handler:
		return composeMiddlewareAdapter(mw)
	case func(Handler) Handler:
		return mw
	case func(http.Handler) http.Handler:
		return httpMiddlewareAdapter(mw)
	default:
		h, err := convertHandler(mw)
		if err != nil {
			panic(fmt.Sprintf("Unsupported middleware type '%v' at index %d.", reflect.TypeOf(middleware[i]), i))
		}
		return handlerMiddlewareAdapter(h)
	}
}

// convert middleware adapters

func composeMiddlewareAdapter(mw func(interface{}) Handler) func(Handler) Handler {
	return func(next Handler) Handler { return mw(next) }
}

func handlerMiddlewareAdapter(mw Handler) func(Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(c *Context) {
			mw.ServeHTTP(c)
			next.ServeHTTP(c)
		})
	}
}

func httpMiddlewareAdapter(mw func(http.Handler) http.Handler) func(Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(c *Context) {
			mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Response = w
				c.Request = r
				next.ServeHTTP(c)
			})).ServeHTTP(c.Response, c.Request)
		})
	}
}

// handler

func convertHandlerPanic(v interface{}) Handler {
	h, err := convertHandler(v)
	if err != nil {
		panic(err)
	}
	return h
}

func convertHandler(v interface{}) (Handler, error) {
	if v == nil {
		return nil, fmt.Errorf("Handler of type '%v' is nil.", reflect.TypeOf(v))
	}
	switch h := v.(type) {
	case Handler: // HandlerFunc, HandlerSpreadFunc implements Handler
		return h, nil
	case func(*Context):
		return HandlerFunc(h), nil
	case func(context.Context, http.ResponseWriter, *http.Request):
		return handlerSpreadAdapter(h), nil
	case http.Handler: // http.HandlerFunc implements http.Handler
		return httpHandlerAdapter(h), nil
	case func(http.ResponseWriter, *http.Request):
		return httpHandlerAdapter(http.HandlerFunc(h)), nil
	case ResultHandler: // ResultHandlerFunc, ResultHandlerSpreadFunc implements ResultHandler
		return resultHandlerAdapter(h), nil
	case func(*Context) Handler:
		return resultHandlerAdapter(ResultHandlerFunc(h)), nil
	case func(context.Context, http.ResponseWriter, *http.Request) Handler:
		return resultHandlerSpreadAdapter(h), nil
	default:
		return nil, fmt.Errorf("Unsupported handler type '%v'", reflect.TypeOf(h))
	}
}

// convert handler adapters

func handlerSpreadAdapter(h func(c context.Context, w http.ResponseWriter, r *http.Request)) Handler {
	return HandlerFunc(func(c *Context) {
		h(c.Context, c.Response, c.Request)
	})
}

func httpHandlerAdapter(h http.Handler) Handler {
	return HandlerFunc(func(c *Context) {
		h.ServeHTTP(c.Response, c.Request)
	})
}

func resultHandlerAdapter(h ResultHandler) Handler {
	return HandlerFunc(func(c *Context) {
		h.ServeHTTP(c).ServeHTTP(c)
	})
}

func resultHandlerSpreadAdapter(h func(context.Context, http.ResponseWriter, *http.Request) Handler) Handler {
	return HandlerFunc(func(c *Context) {
		h(c.Context, c.Response, c.Request).ServeHTTP(c)
	})
}
