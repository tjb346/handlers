package handlers

import (
	"io/ioutil"
	"net/http"
)

var SERIALIZATION_ERROR_MESSAGE = "Could not parse request body."

type ListView interface {
	GetDefault() interface{}
	SaveObject(object interface{})
	GetObjects(r *http.Request) []interface{}
	GetSerializer(contentType string) Serializer
	Validate(object interface{}) error
}

type ObjectView interface {
	GetDefault() interface{}
	GetObject(r *http.Request) interface{}
	SaveObject(object interface{})
	DeleteObject(object interface{})
	GetSerializer(contentType string) ObjectSerializer
	Validate(object interface{}) error
}

func CreateListHandler(view ListView) http.Handler {
	return HTTPMethodHandler{
		GET: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			objects := view.GetObjects(r)
			contentType := r.Header.Get("Content-Type")
			resourceJSON, err := view.GetSerializer(contentType).SerializeList(objects)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(resourceJSON)
		}),

		POST: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			object := view.GetDefault()
			contentType := r.Header.Get("Content-Type")
			serializer := view.GetSerializer(contentType)

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if serializer.DeserializeObject(body, object) != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(SERIALIZATION_ERROR_MESSAGE))
				return
			}

			validationError := view.Validate(object)
			if validationError != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(validationError.Error()))
				return
			}

			view.SaveObject(object)

			serializedObject, serializerErr := serializer.SerializeObject(object)
			if serializerErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			w.Write(serializedObject)
		}),
	}
}

func CreateObjectHandler(view ObjectView) http.Handler {
	return HTTPMethodHandler{
		GET: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			object := view.GetObject(r)
			if object == nil {
				w.WriteHeader(http.StatusNotFound)
			}
			contentType := r.Header.Get("Content-Type")
			serializedObject, serializerErr := view.GetSerializer(contentType).SerializeObject(object)
			if serializerErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(serializedObject)
		}),

		PATCH: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			object := view.GetObject(r)
			contentType := r.Header.Get("Content-Type")
			serializer := view.GetSerializer(contentType)

			if object == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err = serializer.DeserializeObject(body, object)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(SERIALIZATION_ERROR_MESSAGE))
				return
			}

			validationError := view.Validate(object)
			if validationError != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(validationError.Error()))
				return
			}

			view.SaveObject(object)

			serializedObject, serializerErr := serializer.SerializeObject(object)
			if serializerErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(serializedObject)
		}),

		PUT: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			object := view.GetDefault()
			contentType := r.Header.Get("Content-Type")
			serializer := view.GetSerializer(contentType)

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err = serializer.DeserializeObject(body, object)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(SERIALIZATION_ERROR_MESSAGE))
				return
			}

			validationError := view.Validate(object)
			if validationError != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(validationError.Error()))
				return
			}

			view.SaveObject(object)

			serializedObject, serializerErr := serializer.SerializeObject(object)
			if serializerErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			w.Write(serializedObject)
		}),

		DELETE: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			object := view.GetObject(r)
			if object == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			view.DeleteObject(object)
			w.WriteHeader(http.StatusOK)
		}),
	}
}
