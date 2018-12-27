package handlers

import "encoding/json"

type Factory interface {
	Create() Persistable
}

type Persistable interface {
	Reset()
	Save() FieldErrors
	Delete()
}

type JSONResource struct {
	object Persistable
}

func (resource *JSONResource) GetContentType() string {
	return "application/json"
}

func (resource *JSONResource) Read() ([]byte, error) {
	return json.Marshal(resource.object)
}

func (resource *JSONResource) Update(data []byte) error {
	resource.object.Reset()
	return resource.PartialUpdate(data)
}

func (resource *JSONResource) PartialUpdate(data []byte) error {
	err := json.Unmarshal(data, &resource.object)
	if err != nil {
		return err
	}
	fieldErrs := resource.object.Save()
	if fieldErrs != nil {
		return fieldErrs
	}
	return nil
}

func (resource JSONResource) Delete() {
	resource.object.Delete()
}

type JSONListResource struct {
	objectList []interface{}
	creator    Factory
}

func (resource *JSONListResource) GetContentType() string {
	return "application/json"
}

func (resource *JSONListResource) Read() ([]byte, error) {
	return json.Marshal(resource.objectList)
}

func (resource *JSONListResource) Create(data []byte) (Readable, error) {
	newObj := resource.creator.Create()
	err := json.Unmarshal(data, &newObj)
	if err != nil {
		return nil, err
	}
	fieldErrs := newObj.Save()
	if fieldErrs != nil {
		return nil, fieldErrs
	}
	return &JSONResource{object: newObj}, nil
}
