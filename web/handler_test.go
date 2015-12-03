package web_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	web "github.com/shelakel/go-exp-mvc/web"
)

var _ = Describe("ComposeMiddleware", func() {

	It("should convert compatible middleware", func() {
		i := 0
		middleware := func(next web.Handler) web.Handler { i++; return next }
		handlerFuncSignature := func(c web.Context, w http.ResponseWriter, r *http.Request) { i++ }
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
			ServeHTTP(nil, nil, nil)
		Expect(i).To(Equal(9))
	})

	It("should convert compatible handlers", func() {
		i := 0
		handlerFuncSignature := func(c web.Context, w http.ResponseWriter, r *http.Request) { i++ }
		handlerFunc := web.HandlerFunc(handlerFuncSignature)
		handler := web.Handler(handlerFunc)
		httpHandlerFuncSignature := func(w http.ResponseWriter, r *http.Request) { i++ }
		httpHandlerFunc := http.HandlerFunc(httpHandlerFuncSignature)
		httpHandler := http.Handler(httpHandlerFunc)
		handlers := []interface{}{
			handler,
			handlerFunc,
			handlerFuncSignature,
			httpHandler,
			httpHandlerFunc,
			httpHandlerFuncSignature,
		}
		for _, h := range handlers {
			web.ComposeMiddleware()(h).ServeHTTP(nil, nil, nil)
		}
		Expect(i).To(Equal(6))
	})

})
