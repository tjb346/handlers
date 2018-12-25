package handlers

import "encoding/json"

type ObjectSerializer interface {
	SerializeObject(instance interface{}) ([]byte, error)
	DeserializeObject(data []byte, instance interface{}) error
}

type ListSerializer interface {
	SerializeList(instances []interface{}) ([]byte, error)
	DeserializeList(data []byte, instances []interface{}) error
}

type Serializer interface {
	ObjectSerializer
	ListSerializer
}

type JSONSerializer struct{}

func (serializer JSONSerializer) SerializeObject(instance interface{}) ([]byte, error) {
	return json.Marshal(instance)
}

func (serializer JSONSerializer) DeserializeObject(data []byte, instance interface{}) error {
	err := json.Unmarshal(data, instance)
	if err != nil {
		return err
	}
	return nil
}

func (serializer JSONSerializer) SerializeList(instances []interface{}) ([]byte, error) {
	return serializer.SerializeObject(instances)
}

func (serializer JSONSerializer) DeserializeList(data []byte, instances []interface{}) error {
	return serializer.DeserializeObject(data, instances)
}
