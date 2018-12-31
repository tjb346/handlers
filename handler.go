package handlers

import (
	"net/http"
)

// Represents a REST endpoint. Should return the appropriate Resource for the given request. For
// instance you may get the id of an object for the request and the content type from
// the Accept header and return a resource that will serialize that object with the content type.
type Endpoint interface {
	GetResource(r *http.Request) Resource
}

// Implements a handler for a REST endpoint given the resource dispatcher.
type EndpointHandler struct {
	Endpoint Endpoint
}

func (handler EndpointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resource := handler.Endpoint.GetResource(r)
	if resource == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", resource.GetContentType())
	methodHandler := HTTPMethodHandler{
		dispatcher: restHandlerDispatcher{resource: resource},
	}
	methodHandler.ServeHTTP(w, r)
}
