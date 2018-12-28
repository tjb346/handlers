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
	Delete()
}

// A REST endpoint. Should return the appropriate Resource for the given request. For
// instance you may get the id of an object for the request and the content type from
// the Accept header and return a resource that will serialize that object with the content type.
type Endpoint interface {
	GetResource(r *http.Request) Resource
}

func CreateHandler(endpoint Endpoint) http.Handler {
	dispatcher := HTTPMethodDispatcher{
		GET: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resource := endpoint.GetResource(r)
			if resource == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			readable, isReadable := resource.(Readable)
			if !isReadable {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			data, err := readable.Read()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}),
		POST: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resource := endpoint.GetResource(r)
			if resource == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			creatable, isCreatable := resource.(Creatable)
			if !isCreatable {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			newReadable, err := creatable.Create(body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, fieldErrOk := err.(FieldErrors)
				if fieldErrOk {
					w.Header().Set("Content-Type", "application/json")
				}
				w.Write([]byte(err.Error()))
				return
			}

			data, err := newReadable.Read()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", resource.GetContentType())
			w.WriteHeader(http.StatusCreated)
			w.Write(data)
		}),
		PATCH: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resource := endpoint.GetResource(r)
			if resource == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			partialUpdatable, isPartialUpdatable := resource.(PartialUpdatable)
			if !isPartialUpdatable {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err = partialUpdatable.PartialUpdate(body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, fieldErrOk := err.(FieldErrors)
				if fieldErrOk {
					w.Header().Set("Content-Type", "application/json")
				}
				w.Write([]byte(err.Error()))
				return
			}

			readable, isReadable := resource.(Readable)
			if isReadable {
				data, err := readable.Read()
				if err == nil {
					w.Header().Set("Content-Type", resource.GetContentType())
					w.Write(data)
				}
			}

			w.WriteHeader(http.StatusOK)

		}),
		PUT: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resource := endpoint.GetResource(r)
			if resource == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			updatable, isUpdatable := resource.(Updatable)
			if !isUpdatable {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err = updatable.Update(body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, fieldErrOk := err.(FieldErrors)
				if fieldErrOk {
					w.Header().Set("Content-Type", "application/json")
				}
				w.Write([]byte(err.Error()))
				return
			}

			readable, isReadable := resource.(Readable)
			if isReadable {
				data, err := readable.Read()
				if err == nil {
					w.Header().Set("Content-Type", resource.GetContentType())
					w.Write(data)
				}
			}

			w.WriteHeader(http.StatusOK)
		}),
		DELETE: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			readable := endpoint.GetResource(r)
			if readable == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			deletable, isDeletable := readable.(Deletable)
			if !isDeletable {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			deletable.Delete()
			w.WriteHeader(http.StatusOK)
		}),
	}

	return HTTPMethodHandler{dispatcher: dispatcher}
}
