package handlers

import (
	"encoding/json"
	"errors"
)

const jsonContentType = "application/json"

type Savable interface {
	Save() error
}

type Validatable interface {
	Validate() FieldErrors
}

type SafeSavable interface {
	Savable
	Validatable
}

// Return the default new object. The data in POST requests will be added to this"
type Factory interface {
	Create() SafeSavable
}

type ResourceObject interface {
	SafeSavable
	Deletable
	Reset()
}

// An object resource for creating a simple readonly JSON REST endpoint. Will render the given
// object using JSON marshall for any GET request.
type JSONReadOnlyResource struct {
	Object interface{}
}

func (resource *JSONReadOnlyResource) GetContentType() string {
	return jsonContentType
}

func (resource *JSONReadOnlyResource) Read() ([]byte, error) {
	return json.Marshal(resource.Object)
}

// An object resource for creating a simple JSON REST endpoint. Will render the given
// object using json.Marshall for any GET request. Will also allow for PUT, PATCH,
// operations using json.Unmarshall. Also allows for DELETE operations.
type JSONResource struct {
	Object ResourceObject
}

func (resource *JSONResource) GetContentType() string {
	return jsonContentType
}

func (resource *JSONResource) Read() ([]byte, error) {
	return json.Marshal(resource.Object)
}

func (resource *JSONResource) Update(data []byte) error {
	resource.Object.Reset()
	return resource.PartialUpdate(data)
}

func (resource *JSONResource) PartialUpdate(data []byte) error {
	err := json.Unmarshal(data, &resource.Object)
	if err != nil {
		return err
	}
	fieldErrs := resource.Object.Validate()
	if fieldErrs != nil {
		return fieldErrs
	}
	return resource.Object.Save()
}

func (resource JSONResource) Delete() error {
	return resource.Object.Delete()
}

// A list resource that will return a JSON array of the given ObjectList for
// a GET request.
type JSONReadOnlyListResource struct {
	ObjectList []interface{}
}

func (resource *JSONReadOnlyListResource) GetContentType() string {
	return jsonContentType
}

func (resource *JSONReadOnlyListResource) Read() ([]byte, error) {
	return json.Marshal(resource.ObjectList)
}

// A list resource that will return a JSON array of the given ObjectList for
// a GET request. Will create objects on a POST request using json.Unmarshall on
// the default object created by the Creator Factory.
type JSONListResource struct {
	ObjectList []interface{}
	Creator    Factory
}

func (resource *JSONListResource) GetContentType() string {
	return jsonContentType
}

func (resource *JSONListResource) Read() ([]byte, error) {
	return json.Marshal(resource.ObjectList)
}

func (resource *JSONListResource) Create(data []byte) (Readable, error) {
	newObj := resource.Creator.Create()
	err := json.Unmarshal(data, &newObj)
	if err != nil {
		return nil, err
	}
	fieldErrs := newObj.Validate()
	if fieldErrs != nil {
		return nil, fieldErrs
	}
	err = newObj.Save()
	if err != nil {
		return nil, errors.New("error saving new object")
	}
	return &JSONReadOnlyResource{Object: newObj}, nil
}
