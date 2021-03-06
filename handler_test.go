package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

var dataStore = make(map[string]*PetObject)

type PetObject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json: "age,omitempty"`
}

func (pet *PetObject) Validate() FieldErrors {
	if pet.Age > 10 {
		errs := NewFieldErrors()
		errs.Add("age", "Too old")
		return errs
	}
	return nil
}

func (pet *PetObject) GetContentType() string {
	return "application/json"
}

func (pet *PetObject) Read() ([]byte, error) {
	return json.Marshal(pet)
}

func (pet *PetObject) Update(data []byte) error {
	newPet := PetObject{
		ID: pet.ID,
	}
	err := json.Unmarshal(data, &newPet)
	if err != nil {
		return err
	}
	fieldErrs := newPet.Validate()
	if fieldErrs == nil {
		dataStore[pet.ID] = &newPet
	}
	return fieldErrs
}

func (pet *PetObject) PartialUpdate(data []byte) error {
	err := json.Unmarshal(data, pet)
	if err != nil {
		return err
	}
	fieldErrs := pet.Validate()
	if fieldErrs == nil {
		dataStore[pet.ID] = pet
	}
	return fieldErrs
}

func (pet *PetObject) Delete() error {
	delete(dataStore, pet.ID)
	return nil
}

type PetObjectResourceDispatcher struct{}

func (endpoint PetObjectResourceDispatcher) GetResource(r *http.Request) Resource {
	id := strings.Trim(r.URL.Path, "/")
	pet := dataStore[id]

	if pet == nil {
		return nil
	}

	return pet
}

var PetObjectEndpoint = EndpointHandler{Endpoint: PetObjectResourceDispatcher{}}

type PetList struct {
}

func (pet *PetList) GetContentType() string {
	return "application/json"
}

func (petList *PetList) Read() ([]byte, error) {
	pets := make([]*PetObject, 0)
	for _, value := range dataStore {
		pets = append(pets, value)
	}
	return json.Marshal(pets)
}

func (petList *PetList) Create(data []byte) (Readable, error) {
	pet := PetObject{}
	if dataStore[pet.ID] != nil {
		fieldErrs := NewFieldErrors()
		fieldErrs.Add("id", "Field with id already exists.")
		return nil, fieldErrs
	}
	err := json.Unmarshal(data, &pet)
	if err != nil {
		return nil, err
	}

	err = pet.Validate()
	if err != nil {
		return nil, err
	}

	dataStore[pet.ID] = &pet
	return &pet, nil
}

type PetListResourceDispatcher struct{}

func (endpoint PetListResourceDispatcher) GetResource(r *http.Request) Resource {
	return &PetList{}
}

var PetListHandler = EndpointHandler{Endpoint: PetListResourceDispatcher{}}

func TestGetList(t *testing.T) {
	dataStore = make(map[string]*PetObject) // Empty data store
	pet1 := PetObject{
		ID:   "foo",
		Name: "Foo",
		Age:  5,
	}
	dataStore["foo"] = &pet1
	pet2 := PetObject{
		ID:   "bar",
		Name: "Bar",
		Age:  5,
	}
	dataStore["bar"] = &pet2

	validRequest := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	w := httptest.NewRecorder()
	PetListHandler.ServeHTTP(w, validRequest)
	if w.Code != http.StatusOK {
		t.Error("Should return a 200.")
	}
	if w.Header().Get("Content-Type") != (&PetList{}).GetContentType() {
		t.Error("Wrong content type returned. Should be equal to Resource.GetContentType()")
	}
	var petObjects []PetObject
	jsonErr := json.Unmarshal(w.Body.Bytes(), &petObjects)
	if jsonErr != nil {
		t.Error("Returned invalid json.")
	}

	if len(petObjects) != 2 {
		t.Error("Did not get all 2 objects. Only got " + strconv.Itoa(len(petObjects)) + ".")
	}
}

func TestGetObject(t *testing.T) {
	dataStore = make(map[string]*PetObject) // Empty data store
	id := "foo"
	pet := PetObject{
		ID:   id,
		Name: "Foo",
		Age:  5,
	}
	dataStore[id] = &pet

	badIdRequest := httptest.NewRequest(http.MethodGet, "http://example.com/wrongId", nil)
	validRequest := httptest.NewRequest(http.MethodGet, "http://example.com/"+id, nil)

	w := httptest.NewRecorder()
	PetObjectEndpoint.ServeHTTP(w, badIdRequest)
	if w.Code != http.StatusNotFound {
		t.Error("Should not find a resource that has not been created.")
	}

	w = httptest.NewRecorder()
	PetObjectEndpoint.ServeHTTP(w, validRequest)
	if w.Code != http.StatusOK {
		t.Error("Should be able to get created resource.")
	}
	if w.Header().Get("Content-Type") != pet.GetContentType() {
		t.Error("Wrong content type returned. Should be equal to Resource.GetContentType()")
	}
}

func TestCreateObject(t *testing.T) {
	dataStore = make(map[string]*PetObject) // Empty data store
	id := "foo"
	name := "jinx"

	validJSON := "{\"id\":\"" + id + "\",\"name\":\"" + name + "\", \"age\": 5}"
	invalidJSON := "{\"id\":\"" + id + "\",\"name\":\"" + name + "\", \"age\": 25"  // Missing closing bracket
	invalidData := "{\"id\":\"" + id + "\",\"name\":\"" + name + "\", \"age\": 25}" // Too old

	validRequest := httptest.NewRequest(http.MethodPost, "http://example.com/", strings.NewReader(validJSON))
	invalidJSONRequest := httptest.NewRequest(http.MethodPost, "http://example.com/", strings.NewReader(invalidJSON)) // Missing closing bracket
	invalidDataRequest := httptest.NewRequest(http.MethodPost, "http://example.com/", strings.NewReader(invalidData)) // Too old

	w := httptest.NewRecorder()
	PetListHandler.ServeHTTP(w, invalidJSONRequest)
	if w.Code != http.StatusBadRequest {
		t.Error("Invalid response code " + strconv.Itoa(w.Code) + " should 400.")
	}
	if w.Body.String() != "unexpected end of JSON input" {
		t.Error("Wrong error message returned. Returned " + w.Body.String())
	}

	w = httptest.NewRecorder()
	PetListHandler.ServeHTTP(w, invalidDataRequest)
	if w.Code != http.StatusBadRequest {
		t.Error("Invalid response code " + strconv.Itoa(w.Code) + " should 400.")
	}

	w = httptest.NewRecorder()
	PetListHandler.ServeHTTP(w, validRequest)
	if w.Code != http.StatusCreated {
		t.Error("Invalid response code " + strconv.Itoa(w.Code) + " should 201.")
	}
	newObj := PetObject{}
	jsonErr := json.Unmarshal(w.Body.Bytes(), &newObj)
	if jsonErr != nil {
		t.Error("Returned invalid json.")
	}
	if newObj.ID != id {
		t.Error("Returned wrong ObjectList.")
	}

	savedPet := dataStore[id]
	if savedPet == nil {
		t.Error("Value not saved")
	}
	if savedPet.Name != name {
		t.Error("Name not saved")
	}
}

func TestPatchObject(t *testing.T) {
	dataStore = make(map[string]*PetObject)
	id := "bar"
	name := "buck"
	newAge := 7
	pet := PetObject{
		ID:   id,
		Name: name,
		Age:  5,
	}
	dataStore[id] = &pet

	patchAgeJSON := "{\"age\": " + strconv.Itoa(newAge) + "}"
	invalidData := "{\"age\": 25}" // Too old

	wrongResourceRequest := httptest.NewRequest(http.MethodPatch, "http://example.com/wrongId", strings.NewReader(patchAgeJSON))
	validRequest := httptest.NewRequest(http.MethodPatch, "http://example.com/"+id, strings.NewReader(patchAgeJSON))
	inValidDataRequest := httptest.NewRequest(http.MethodPatch, "http://example.com/"+id, strings.NewReader(invalidData))

	w := httptest.NewRecorder()
	PetListHandler.ServeHTTP(w, validRequest)
	if w.Code != http.StatusMethodNotAllowed {
		t.Error("Invalid response code " + strconv.Itoa(w.Code) + " should 405. PetList does not implement Updatable")
	}

	w = httptest.NewRecorder()
	PetObjectEndpoint.ServeHTTP(w, wrongResourceRequest)
	if w.Code != http.StatusNotFound {
		t.Error("Invalid response code " + strconv.Itoa(w.Code) + " should 404.")
	}

	w = httptest.NewRecorder()
	PetObjectEndpoint.ServeHTTP(w, validRequest)
	if w.Code != http.StatusOK {
		t.Error("Should be able to patch resource.")
	}
	if w.Header().Get("Content-Type") != (&PetList{}).GetContentType() {
		t.Error("Wrong content type returned. Should be equal to Resource.GetContentType()")
	}
	newObj := PetObject{}
	jsonErr := json.Unmarshal(w.Body.Bytes(), &newObj)
	if jsonErr != nil {
		t.Error("Returned invalid json.")
	}
	if newObj.ID != id {
		t.Error("Wrong value returned.")
	}
	newObj = *dataStore[id]
	if newObj.Age != newAge {
		t.Error("Age not patched.")
	}
	if newObj.Name != name {
		t.Error("Name changed.")
	}

	w = httptest.NewRecorder()
	PetObjectEndpoint.ServeHTTP(w, inValidDataRequest)
	if w.Code != http.StatusBadRequest {
		t.Error("Should not allow patch with bad data.")
	}
}

func TestDeleteObject(t *testing.T) {
	dataStore = make(map[string]*PetObject)
	id := "bar"
	pet := PetObject{
		ID:   id,
		Name: "Foo",
		Age:  5,
	}
	dataStore[id] = &pet

	wrongResourceRequest := httptest.NewRequest(http.MethodDelete, "http://example.com/wrongId", nil)
	validRequest := httptest.NewRequest(http.MethodDelete, "http://example.com/"+id, nil)

	w := httptest.NewRecorder()
	PetListHandler.ServeHTTP(w, validRequest)
	if w.Code != http.StatusMethodNotAllowed {
		t.Error("Invalid response code " + strconv.Itoa(w.Code) + " should 405. PetList does not implement Deletable")
	}

	w = httptest.NewRecorder()
	PetObjectEndpoint.ServeHTTP(w, wrongResourceRequest)
	if w.Code != http.StatusNotFound {
		t.Error("Should return not found.")
	}

	w = httptest.NewRecorder()
	PetObjectEndpoint.ServeHTTP(w, validRequest)
	if w.Code != http.StatusOK {
		t.Error("Should be able to delete created resource.")
	}
}
