package handlers

import "net/http"

type HTTPMethods interface {
	GET(w http.ResponseWriter, r *http.Request)
	POST(w http.ResponseWriter, r *http.Request)
	PATCH(w http.ResponseWriter, r *http.Request)
	PUT(w http.ResponseWriter, r *http.Request)
	DELETE(w http.ResponseWriter, r *http.Request)
	OPTIONS(w http.ResponseWriter, r *http.Request)
}

type HTTPMethodHandler struct {
	GET     http.Handler
	POST    http.Handler
	PATCH   http.Handler
	PUT     http.Handler
	DELETE  http.Handler
	OPTIONS http.Handler
}

func (handler HTTPMethodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var method http.Handler
	switch r.Method {
	case http.MethodGet:
		method = handler.GET
		break
	case http.MethodPost:
		method = handler.POST
		break
	case http.MethodPatch:
		method = handler.PATCH
		break
	case http.MethodPut:
		method = handler.PUT
		break
	case http.MethodDelete:
		method = handler.DELETE
		break
	case http.MethodOptions:
		method = handler.OPTIONS
		break
	default:
		break
	}

	if method == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		method.ServeHTTP(w, r)
	}
}

func CreateHandler(methods HTTPMethods) HTTPMethodHandler {
	return HTTPMethodHandler{
		GET:    http.HandlerFunc(methods.GET),
		POST:   http.HandlerFunc(methods.POST),
		PUT:    http.HandlerFunc(methods.PUT),
		PATCH:  http.HandlerFunc(methods.PATCH),
		DELETE: http.HandlerFunc(methods.DELETE),
	}
}

func CreateReadOnlyHandler(methods HTTPMethods) HTTPMethodHandler {
	return HTTPMethodHandler{
		GET: http.HandlerFunc(methods.GET),
	}
}
