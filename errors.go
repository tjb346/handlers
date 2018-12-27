package handlers

import "encoding/json"

type FieldErrors interface {
	error
	json.Marshaler
	Add(field string, errorMessage string)
}

type fieldErrors struct {
	errors map[string]string
}

func (fieldErrors *fieldErrors) MarshalJSON() ([]byte, error) {
	if fieldErrors.errors == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(fieldErrors.errors)
}

func (fieldErrors *fieldErrors) Add(field string, errorMessage string) {
	if fieldErrors.errors == nil {
		fieldErrors.errors = make(map[string]string)
	}
	fieldErrors.errors[field] = errorMessage
}

func (fieldErrors *fieldErrors) Error() string {
	data, err := json.Marshal(fieldErrors)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func NewFieldErrors() FieldErrors {
	return &fieldErrors{errors: make(map[string]string)}
}
