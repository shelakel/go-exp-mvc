package web

import (
	"net/http"

	"encoding/json"
	"encoding/xml"

	"github.com/shelakel/go-exp-mvc/web/header"
)

// An ActionResult in the traditional sense is a func that returns a Handler.

// Result renders v using content negotiation with a status code.
func Result(v interface{}, code int) Handler {
	return contentNegotiationResult(v, http.StatusOK)
}

var accepted = []string{
	"application/json", "text/json",
	"application/xml", "text/xml",
}

func contentNegotiationResult(v interface{}, code int) Handler {
	return HandlerFunc(func(c *Context) {
		w := c.Response
		if v == nil {
			w.WriteHeader(code)
			return
		}
		var (
			ok     bool
			err    error
			data   []byte
			accept string
		)
		if err, ok = v.(error); ok {
			panic(err)
		}
		accept = header.NegotiateContentType(c.Request, accepted, "application/json")
		switch accept {
		case "application/json":
			data, err = json.Marshal(v)
		case "text/json":
			data, err = json.MarshalIndent(v, "", "    ")
		case "application/xml":
			data, err = xml.Marshal(v)
		case "text/xml":
			data, err = xml.MarshalIndent(v, "", "    ")
		}
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", accept+";charset=utf-8")
		w.WriteHeader(code)
		w.Write(data) // ignore err?
	})
}
