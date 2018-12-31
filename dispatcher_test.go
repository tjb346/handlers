package handlers

import (
	"net/http"
	"testing"
)

type resource struct{}

func (resource resource) GetContentType() string {
	return "text/plain"
}

type readable struct {
	resource
}

func (readable readable) Read() ([]byte, error) {
	return []byte{}, nil
}

type creatable struct {
	resource
}

func (creatable creatable) Create(data []byte) (Readable, error) {
	return readable{}, nil
}

type deletable struct {
	resource
}

func (deletable deletable) Delete() {
}

func TestRESTDispatcher(t *testing.T) {
	dispatcher := restHandlerDispatcher{resource: resource{}}
	if dispatcher.GetMethodHandler(http.MethodGet) != nil {
		t.Error("Should return nil for non-Readable resource")
	}
	if dispatcher.GetMethodHandler(http.MethodPost) != nil {
		t.Error("Should return nil for non-Creatable resource")
	}
	if dispatcher.GetMethodHandler(http.MethodPatch) != nil {
		t.Error("Should return nil for non-PartialUpdatable resource")
	}
	if dispatcher.GetMethodHandler(http.MethodPut) != nil {
		t.Error("Should return nil for non-Updatable resource")
	}
	if dispatcher.GetMethodHandler(http.MethodDelete) != nil {
		t.Error("Should return nil for non-Deletable resource")
	}

	dispatcher = restHandlerDispatcher{resource: readable{}}
	if dispatcher.GetMethodHandler(http.MethodGet) == nil {
		t.Error("Should not return nil for Readable resource")
	}
	dispatcher = restHandlerDispatcher{resource: creatable{}}
	if dispatcher.GetMethodHandler(http.MethodPost) == nil {
		t.Error("Should not return nil for Creatable resource")
	}
	dispatcher = restHandlerDispatcher{resource: deletable{}}
	if dispatcher.GetMethodHandler(http.MethodDelete) == nil {
		t.Error("Should not return nil for Deletable resource")
	}
}
