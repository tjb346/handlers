package handlers

import (
	"encoding/json"
	"testing"
)

var DisallowedNames = []string{"Fred"}

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (person *Person) Reset() {
	person.Name = ""
	person.Age = 0
}

func (person *Person) Save() FieldErrors {
	validName := true
	for _, invalidName := range DisallowedNames {
		if person.Name == invalidName {
			validName = false
		}
	}
	if !validName {
		errs := NewFieldErrors()
		errs.Add("name", "invalid")
		return errs
	}
	people.store = append(people.store, *person)
	return nil
}

func (person *Person) Delete() {
	newPeople := make([]Person, 0)
	for _, person := range people.store {
		if person.Name != person.Name {
			newPeople = append(newPeople, person)
		}
	}
	people.store = newPeople
}

type People struct {
	store []Person
}

func (people *People) Create() Persistable {
	return &Person{}
}

func (people *People) InterfaceList() []interface{} {
	interfaceList := make([]interface{}, len(people.store))
	for i, person := range people.store {
		interfaceList[i] = person
	}
	return interfaceList
}

var people People

func TestGenericJsonResource(t *testing.T) {
	person := Person{
		Name: "Bob",
		Age:  35,
	}
	personResource := JSONResource{
		object: &person,
	}

	data, err := personResource.Read()
	if err != nil {
		t.Error("Error reading data.")
	}
	returnedPerson := Person{}
	err = json.Unmarshal(data, &returnedPerson)
	if err != nil {
		t.Error("Returned invalid json.")
	}
	if returnedPerson.Name != person.Name {
		t.Error("Wrong person or person data returned.")
	}
}

func TestGenericJsonListResource(t *testing.T) {
	person1 := Person{
		Name: "Bob",
		Age:  35,
	}
	person2 := Person{
		Name: "Jim",
		Age:  43,
	}

	people = People{store: []Person{person1, person2}}

	peopleResource := JSONListResource{
		objectList: people.InterfaceList(),
		creator:    &people,
	}

	data, err := peopleResource.Read()
	if err != nil {
		t.Error("Error reading data.")
	}
	returnedPeople := make([]Person, 0)
	err = json.Unmarshal(data, &returnedPeople)
	if err != nil {
		t.Error("Returned invalid json.")
	}
	if len(returnedPeople) != len(people.store) {
		t.Error("Wrong number of people returned.")
	}

	newPerson := Person{Name: "Fred"}
	data, err = json.Marshal(&newPerson)
	if err != nil {
		t.Error("Json test err.")
	}
	_, errs := peopleResource.Create(data)
	if errs == nil {
		t.Error("Invalid name should error.")
	}

	newPerson = Person{Name: "Dave"}
	data, err = json.Marshal(&newPerson)
	if err != nil {
		t.Error("Json test err.")
	}
	_, errs = peopleResource.Create(data)
	if errs != nil {
		t.Error("Valid name should not error.")
	}
	if len(people.store) != 3 {
		t.Error("Person not added.")
	}
}
