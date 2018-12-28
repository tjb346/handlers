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

func TestGenericJsonReadonlyResource(t *testing.T) {
	person := Person{
		Name: "Bob",
		Age:  35,
	}
	var personResource Resource
	personResource = &JSONReadOnlyResource{
		Object: &person,
	}

	personReadable, ok := personResource.(Readable)
	if !ok {
		t.Error("JSONResource is not Readable.")
	}

	data, err := personReadable.Read()
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

	_, ok = personResource.(Creatable)
	if ok {
		t.Error("JSONReadOnlyResource should not be Creatable.")
	}

	_, ok = personResource.(PartialUpdatable)
	if ok {
		t.Error("JSONReadOnlyResource should not be PartialUpdatable.")
	}

	_, ok = personResource.(Deletable)
	if ok {
		t.Error("JSONReadOnlyResource should not be Deletable.")
	}
}

func TestGenericJsonResource(t *testing.T) {
	person := Person{
		Name: "Bob",
		Age:  35,
	}
	var personResource Resource
	personResource = &JSONResource{
		Object: &person,
	}

	personReadable, ok := personResource.(Readable)
	if !ok {
		t.Error("JSONResource is not Readable.")
	}

	data, err := personReadable.Read()
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

	personUpdatable, ok := personResource.(PartialUpdatable)
	if !ok {
		t.Error("JSONResource is not PartialUpdatable.")
	}

	newPerson := struct{ Age int }{Age: 20}
	data, err = json.Marshal(&newPerson)
	if err != nil {
		t.Error("Json test err.")
	}
	err = personUpdatable.PartialUpdate(data)
	if err != nil {
		t.Error("Error updating person age.")
	}
	if person.Age != 20 {
		t.Error("Age not updated.")
	}
	if person.Name != "Bob" {
		t.Error("Name incorrectly changed to " + person.Name + ".")
	}

	people.store = []Person{person}
	personDeletable, ok := personResource.(Deletable)
	if !ok {
		t.Error("JSONResource is not Deletable.")
	}
	personDeletable.Delete()
	if len(people.store) != 0 {
		t.Error("Person not deleted.")
	}

	_, ok = personResource.(Creatable)
	if ok {
		t.Error("Person resource should not be Creatable.")
	}
}

func TestGenericJsonReadonlyListResource(t *testing.T) {
	person1 := Person{
		Name: "Bob",
		Age:  35,
	}
	person2 := Person{
		Name: "Jim",
		Age:  43,
	}

	people = People{store: []Person{person1, person2}}

	var peopleResource Resource
	peopleResource = &JSONReadOnlyListResource{
		ObjectList: people.InterfaceList(),
	}

	peopleReadable, ok := peopleResource.(Readable)
	if !ok {
		t.Error("JSONListReadOnlyResource is not Readable.")
	}

	data, err := peopleReadable.Read()
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

	_, ok = peopleReadable.(Creatable)
	if ok {
		t.Error("JSONReadOnlyListResource should not be Creatable.")
	}

	_, ok = peopleReadable.(PartialUpdatable)
	if ok {
		t.Error("JSONReadOnlyListResource should not be PartialUpdatable.")
	}

	_, ok = peopleReadable.(Deletable)
	if ok {
		t.Error("JJSONReadOnlyListResource should not be Deletable.")
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

	var peopleResource Resource
	peopleResource = &JSONListResource{
		ObjectList: people.InterfaceList(),
		Creator:    &people,
	}

	peopleReadable, ok := peopleResource.(Readable)
	if !ok {
		t.Error("JSONReadOnlyResource is not Readable.")
	}

	data, err := peopleReadable.Read()
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

	peopleCreatable, ok := peopleResource.(Creatable)
	if !ok {
		t.Error("JSONReadOnlyResource is not Creatable.")
	}

	newPerson := Person{Name: "Fred"}
	data, err = json.Marshal(&newPerson)
	if err != nil {
		t.Error("Json test err.")
	}
	_, errs := peopleCreatable.Create(data)
	if errs == nil {
		t.Error("Invalid name should error.")
	}

	newPerson = Person{Name: "Dave"}
	data, err = json.Marshal(&newPerson)
	if err != nil {
		t.Error("Json test err.")
	}
	_, errs = peopleCreatable.Create(data)
	if errs != nil {
		t.Error("Valid name should not error.")
	}
	if len(people.store) != 3 {
		t.Error("Person not added.")
	}
}
