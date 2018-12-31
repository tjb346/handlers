package handlers

import "net/http"

type HandlerDispatcher interface {
	GetMethodHandler(requestMethod string) http.Handler
}

type HTTPMethodDispatcher struct {
	GET     http.Handler
	POST    http.Handler
	PATCH   http.Handler
	PUT     http.Handler
	DELETE  http.Handler
	OPTIONS http.Handler
}

func (handler HTTPMethodDispatcher) GetMethodHandler(requestMethod string) http.Handler {
	var method http.Handler
	switch requestMethod {
	case http.MethodGet:
		method = handler.GET
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

	return method
}

type HTTPMethodHandler struct {
	dispatcher HandlerDispatcher
}

func (handler HTTPMethodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := handler.dispatcher.GetMethodHandler(r.Method)
	if method == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	method.ServeHTTP(w, r)
}
