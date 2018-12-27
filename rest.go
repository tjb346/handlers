package handlers

import (
	"io/ioutil"
	"net/http"
)

type Creatable interface {
	Create(data []byte, contentType string) (Readable, error)
}

type Readable interface {
	Read(contentType string) ([]byte, error)
}

type Updatable interface {
	Update(data []byte, contentType string) error
}

type PartialUpdatable interface {
	PartialUpdate(data []byte, contentType string) error
}

type Deletable interface {
	Delete()
}

type Endpoint interface {
	GetReadable(r *http.Request) Readable
}

func CreateHandler(endpoint Endpoint) http.Handler {
	dispatcher := HTTPMethodDispatcher{
		GET: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			readable := endpoint.GetReadable(r)
			if readable == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if readable == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			data, err := readable.Read("application/json")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}),
		POST: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			readable := endpoint.GetReadable(r)
			if readable == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			contentType := r.Header.Get("Content-Type")

			creatable, isCreatable := readable.(Creatable)
			if !isCreatable {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			newReadable, err := creatable.Create(body, contentType)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, fieldErrOk := err.(FieldErrors)
				if fieldErrOk {
					w.Header().Set("Content-Type", "application/json")
				}
				w.Write([]byte(err.Error()))
				return
			}

			data, err := newReadable.Read("application/json")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write(data)
		}),
		PATCH: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			readable := endpoint.GetReadable(r)
			if readable == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			contentType := r.Header.Get("Content-Type")

			partialUpdatable, isPartialUpdatable := readable.(PartialUpdatable)
			if !isPartialUpdatable {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err = partialUpdatable.PartialUpdate(body, contentType)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, fieldErrOk := err.(FieldErrors)
				if fieldErrOk {
					w.Header().Set("Content-Type", "application/json")
				}
				w.Write([]byte(err.Error()))
				return
			}

			data, err := readable.Read("application/json")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}),
		PUT: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			readable := endpoint.GetReadable(r)
			if readable == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			contentType := r.Header.Get("Content-Type")

			updatable, isUpdatable := readable.(Updatable)
			if !isUpdatable {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err = updatable.Update(body, contentType)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, fieldErrOk := err.(FieldErrors)
				if fieldErrOk {
					w.Header().Set("Content-Type", "application/json")
				}
				w.Write([]byte(err.Error()))
				return
			}

			data, err := readable.Read("application/json")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}),
		DELETE: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			readable := endpoint.GetReadable(r)
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
