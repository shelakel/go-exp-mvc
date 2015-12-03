package web

import (
	"fmt"
	"net/http"
	"reflect"

	"golang.org/x/net/context"
)

// Context embeds net/context.Context
type Context interface {
	context.Context
}

// Handler is a http.Handler with a net/context.Context parameter.
type Handler interface {
	ServeHTTP(c Context, w http.ResponseWriter, r *http.Request)
}

// HandlerFunc implements Handler and Middleware.
type HandlerFunc func(c Context, w http.ResponseWriter, r *http.Request)

func (f HandlerFunc) ServeHTTP(c Context, w http.ResponseWriter, r *http.Request) {
	f(c, w, r)
}

// ComposeMiddleware combines Middleware into a single Middleware function
func ComposeMiddleware(middleware ...interface{}) func(interface{}) Handler {
	if middleware == nil || len(middleware) == 0 {
		return convertHandler
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
		next := convertHandler(v)
		for i := len(adapters) - 1; i >= 0; i-- {
			next = adapters[i](next)
		}
		return next
	}
}

func convertMiddleware(middleware []interface{}, i int) func(Handler) Handler {
	switch mw := middleware[i].(type) {
	case func(interface{}) Handler:
		return composeMiddlewareAdapter(mw)
	case func(Handler) Handler:
		return mw
	case Handler:
		return handlerMiddlewareAdapter(mw)
	case HandlerFunc:
		return handlerMiddlewareAdapter(mw)
	case func(Context, http.ResponseWriter, *http.Request):
		return handlerMiddlewareAdapter(HandlerFunc(mw))
	case func(http.Handler) http.Handler:
		return httpMiddlewareAdapter(mw)
	case http.Handler:
		return httpHandlerMiddlewareAdapter(mw)
	case http.HandlerFunc:
		return httpHandlerMiddlewareAdapter(mw)
	case func(http.ResponseWriter, *http.Request):
		return httpHandlerMiddlewareAdapter(http.HandlerFunc(mw))
	default:
		panic(fmt.Sprintf("Unsupported middleware type '%v' at index %d.", reflect.TypeOf(middleware[i]), i))
	}
}

func convertHandler(v interface{}) Handler {
	if v == nil {
		panic(fmt.Sprintf("Handler of type '%v' is nil.", reflect.TypeOf(v)))
	}
	switch h := v.(type) {
	case Handler:
		return h
	case HandlerFunc:
		return h
	case func(Context, http.ResponseWriter, *http.Request):
		return HandlerFunc(h)
	case http.Handler:
		return httpHandlerAdapter(h)
	case http.HandlerFunc:
		return httpHandlerAdapter(h)
	case func(http.ResponseWriter, *http.Request):
		return httpHandlerAdapter(http.HandlerFunc(h))
	default:
		panic(fmt.Sprintf("Unsupported handler type '%v'", reflect.TypeOf(h)))
	}
}

// convert handler adapters

func httpHandlerAdapter(h http.Handler) Handler {
	return HandlerFunc(func(c Context, w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// middleware adapters

func composeMiddlewareAdapter(mw func(interface{}) Handler) func(Handler) Handler {
	return func(next Handler) Handler { return mw(next) }
}

func handlerMiddlewareAdapter(mw Handler) func(Handler) Handler {
	return func(next Handler) Handler {
		if next == nil {
			return mw
		}
		return HandlerFunc(func(c Context, w http.ResponseWriter, r *http.Request) {
			mw.ServeHTTP(c, w, r)
			next.ServeHTTP(c, w, r)
		})
	}
}

func httpHandlerMiddlewareAdapter(mw http.Handler) func(Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(c Context, w http.ResponseWriter, r *http.Request) {
			mw.ServeHTTP(w, r)
			next.ServeHTTP(c, w, r)
		})
	}
}

func httpMiddlewareAdapter(mw func(http.Handler) http.Handler) func(Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(c Context, w http.ResponseWriter, r *http.Request) {
			mw(http.HandlerFunc(func(w1 http.ResponseWriter, r1 *http.Request) {
				next.ServeHTTP(c, w1, r1)
			})).ServeHTTP(w, r)
		})
	}
}
