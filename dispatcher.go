package handlers

import (
	"io/ioutil"
	"net/http"
)

// A basic REST resource. The methods it will respond to are determined by
// which of the below interfaces are implemented.
type Resource interface {
	GetContentType() string
}

// A resource that implements this will respond to POST requests. Use the given
// data to create a new object.
type Creatable interface {
	Create(data []byte) (Readable, error)
}

// A resource that implements this will respond to GET requests. Should return a
// serialized representation of the object as the first return value a format
// that matches the content type returned by GetContentType.
type Readable interface {
	Read() ([]byte, error)
}

// A resource that implements this will respond to PUT requests. Use the given
// data to update the object.
type Updatable interface {
	Update(data []byte) error
}

// A resource that implements this will respond to PATCH requests. Use the given
// data to update the object.
type PartialUpdatable interface {
	PartialUpdate(data []byte) error
}

// A resource that implements this will respond to DELETE requests. The object should
// be deleted when called.
type Deletable interface {
	Delete() error
}

type restHandlerDispatcher struct {
	resource Resource
}

func (dispatcher restHandlerDispatcher) GetMethodHandler(requestMethod string) http.Handler {
	var method http.Handler
	switch requestMethod {
	case http.MethodGet:
		readable, isReadable := dispatcher.resource.(Readable)
		if isReadable {
			method = getHandler{readable: readable}
		}
		break
	case http.MethodPost:
		creatable, isCreatable := dispatcher.resource.(Creatable)
		if isCreatable {
			method = postHandler{creatable: creatable}
		}
		break
	case http.MethodPatch:
		partialUpdatable, isPartialUpdatable := dispatcher.resource.(PartialUpdatable)
		if isPartialUpdatable {
			method = patchHandler{partialUpdatable: partialUpdatable}
		}
		break
	case http.MethodPut:
		updatable, isUpdatable := dispatcher.resource.(Updatable)
		if isUpdatable {
			method = putHandler{updatable: updatable}
		}
		break
	case http.MethodDelete:
		deletable, isDeletable := dispatcher.resource.(Deletable)
		if isDeletable {
			method = deleteHandler{deletable: deletable}
		}
		break
	case http.MethodOptions:
		break
	default:
		break
	}

	return method
}

type getHandler struct {
	readable Readable
}

func (handler getHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := handler.readable.Read()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type postHandler struct {
	creatable Creatable
}

func (handler postHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newReadable, err := handler.creatable.Create(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, fieldErrOk := err.(FieldErrors)
		if fieldErrOk {
			w.Header().Set("Content-Type", "application/json")
		} else {
			w.Header().Set("Content-Type", "text/plain")
		}
		w.Write([]byte(err.Error()))
		return
	}

	data, err := newReadable.Read()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

type patchHandler struct {
	partialUpdatable PartialUpdatable
}

func (handler patchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = handler.partialUpdatable.PartialUpdate(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, fieldErrOk := err.(FieldErrors)
		if fieldErrOk {
			w.Header().Set("Content-Type", "application/json")
		} else {
			w.Header().Set("Content-Type", "text/plain")
		}
		w.Write([]byte(err.Error()))
		return
	}

	readable, isReadable := handler.partialUpdatable.(Readable)
	if isReadable {
		getHandler{readable: readable}.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

type putHandler struct {
	updatable Updatable
}

func (handler putHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = handler.updatable.Update(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, fieldErrOk := err.(FieldErrors)
		if fieldErrOk {
			w.Header().Set("Content-Type", "application/json")
		} else {
			w.Header().Set("Content-Type", "text/plain")
		}
		w.Write([]byte(err.Error()))
		return
	}

	readable, isReadable := handler.updatable.(Readable)
	if isReadable {
		getHandler{readable: readable}.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

type deleteHandler struct {
	deletable Deletable
}

func (handler deleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := handler.deletable.Delete()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error deleting obect"))
		return
	}
	w.WriteHeader(http.StatusOK)
}
