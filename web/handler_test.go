package web_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	web "github.com/shelakel/go-exp-mvc/web"
	"golang.org/x/net/context"
)

var _ = Describe("ComposeMiddleware", func() {

	It("should convert compatible middleware", func() {
		i := 0
		middleware := func(next web.Handler) web.Handler { i++; return next }
		handlerFuncSignature := func(c *web.Context) { i++ }
		handlerFunc := web.HandlerFunc(handlerFuncSignature)
		handler := web.Handler(handlerFunc)
		httpMiddleware := func(next http.Handler) http.Handler { i++; return next }
		httpHandlerFuncSignature := func(w http.ResponseWriter, r *http.Request) { i++ }
		httpHandlerFunc := http.HandlerFunc(httpHandlerFuncSignature)
		httpHandler := http.Handler(httpHandlerFunc)
		web.ComposeMiddleware(
			web.ComposeMiddleware(), // func (interface{}) web.Handler
			middleware,
			handler,
			handlerFunc,
			handlerFuncSignature,
			httpMiddleware,
			httpHandler,
			httpHandlerFunc,
			httpHandlerFuncSignature,
		)(handler).
			ServeHTTP(&web.Context{})
		Expect(i).To(Equal(9))
	})

	It("should panic if a middleware passed is nil", func() {
		act := func() { web.ComposeMiddleware(nil) }
		Expect(act).Should(Panic())
	})

	It("should panic if a middleware passed isn't convertable", func() {
		act := func() { web.ComposeMiddleware(func() {}) }
		Expect(act).Should(Panic())
	})

	It("should convert compatible handlers", func() {
		i := 0
		handlerFuncSignature := func(c *web.Context) { i++ }
		handlerFunc := web.HandlerFunc(handlerFuncSignature)
		handler := web.Handler(handlerFunc)
		handlerSpreadFuncSignature := func(c context.Context, w http.ResponseWriter, r *http.Request) { i++ }
		handlerSpreadFunc := web.HandlerSpreadFunc(handlerSpreadFuncSignature)
		httpHandlerFuncSignature := func(w http.ResponseWriter, r *http.Request) { i++ }
		httpHandlerFunc := http.HandlerFunc(httpHandlerFuncSignature)
		httpHandler := http.Handler(httpHandlerFunc)
		resultHandlerFuncSignature := func(c *web.Context) web.Handler { return handler }
		resultHandlerFunc := web.ResultHandlerFunc(resultHandlerFuncSignature)
		resultHandlerSpreadFuncSignature := func(c context.Context, w http.ResponseWriter, r *http.Request) web.Handler { return handler }
		resultHandlerSpreadFunc := web.ResultHandlerSpreadFunc(resultHandlerSpreadFuncSignature)
		handlers := []interface{}{
			handler,
			handlerFunc,
			handlerFuncSignature,
			handlerSpreadFuncSignature,
			handlerSpreadFunc,
			httpHandler,
			httpHandlerFunc,
			httpHandlerFuncSignature,
			resultHandlerFuncSignature,
			resultHandlerFunc,
			resultHandlerSpreadFuncSignature,
			resultHandlerSpreadFunc,
		}
		for _, h := range handlers {
			web.ComposeMiddleware()(h).
				ServeHTTP(&web.Context{})
		}
		Expect(i).To(Equal(12))
	})

	It("should panic if a handler passed is nil", func() {
		act := func() { web.ComposeMiddleware()(nil) }
		Expect(act).Should(Panic())
	})

	It("should panic if a handler passed isn't convertable", func() {
		act := func() { web.ComposeMiddleware()(func() {}) }
		Expect(act).Should(Panic())
	})

})
