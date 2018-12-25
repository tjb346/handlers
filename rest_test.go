package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type PetObject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json: "age,omitempty"`
}

var dataStore = make(map[string]*PetObject)

type BaseTestView struct {
	maxAge int
}

func (manager BaseTestView) GetDefault() interface{} {
	return &PetObject{}
}

func (manager BaseTestView) SaveObject(object interface{}) {
	dataType := object.(*PetObject)
	dataStore[dataType.ID] = dataType
}

func (manager BaseTestView) Validate(object interface{}) error {
	dataType := object.(*PetObject)
	if dataType.Age > manager.maxAge {
		return errors.New("Too old")
	}
	return nil
}

func (manager BaseTestView) GetSerializer(contentType string) Serializer {
	return JSONSerializer{}
}

type TestListView struct {
	BaseTestView
}

func (manager TestListView) GetObjects(r *http.Request) []interface{} {
	objects := make([]interface{}, 0)
	for id := range dataStore {
		objects = append(objects, dataStore[id])
	}
	return objects
}

type TestObjectView struct {
	BaseTestView
}

func (manager TestObjectView) GetObject(r *http.Request) interface{} {
	id := strings.Trim(r.URL.Path, "/")
	data := dataStore
	value := data[id]

	if value == nil {
		return nil
	}
	return value
}

func (manager TestObjectView) DeleteObject(object interface{}) {
	dataType := object.(*PetObject)
	delete(dataStore, dataType.ID)
}

func (manager TestObjectView) GetSerializer(contentType string) ObjectSerializer {
	return JSONSerializer{}
}

var baseView = BaseTestView{
	maxAge: 10,
}

var objectHandler = CreateObjectHandler(TestObjectView{
	BaseTestView: baseView,
})

var listHandler = CreateListHandler(TestListView{
	BaseTestView: baseView,
})

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
	listHandler.ServeHTTP(w, validRequest)
	if w.Code != http.StatusOK {
		t.Error("Should return a 200.")
	}
	var petObjects []PetObject
	jsonErr := json.Unmarshal(w.Body.Bytes(), &petObjects)
	if jsonErr != nil {
		t.Error("Returned invalid json.")
	}
	if len(petObjects) != 2 {
		t.Error("Did not get all objects.")
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
	objectHandler.ServeHTTP(w, badIdRequest)
	if w.Code != http.StatusNotFound {
		t.Error("Should not find a resource that has not been created.")
	}

	w = httptest.NewRecorder()
	objectHandler.ServeHTTP(w, validRequest)
	if w.Code != http.StatusOK {
		t.Error("Should be able to get created resource.")
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
	listHandler.ServeHTTP(w, invalidJSONRequest)
	if w.Code != http.StatusBadRequest {
		t.Error("Invalid request should 400.")
	}
	if w.Body.String() != SERIALIZATION_ERROR_MESSAGE {
		t.Error("Wrong error message returned.")
	}

	w = httptest.NewRecorder()
	listHandler.ServeHTTP(w, invalidDataRequest)
	if w.Code != http.StatusBadRequest {
		t.Error("Invalid request should 400.")
	}

	w = httptest.NewRecorder()
	listHandler.ServeHTTP(w, validRequest)
	if w.Code != http.StatusCreated {
		t.Error("Created resource should 201.")
	}
	newObj := PetObject{}
	jsonErr := json.Unmarshal(w.Body.Bytes(), &newObj)
	if jsonErr != nil {
		t.Error("Returned invalid json.")
	}
	if newObj.ID != id {
		t.Error("Returned wrong object.")
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
	objectHandler.ServeHTTP(w, wrongResourceRequest)
	if w.Code != http.StatusNotFound {
		t.Error("Should return not found.")
	}

	w = httptest.NewRecorder()
	objectHandler.ServeHTTP(w, validRequest)
	if w.Code != http.StatusOK {
		t.Error("Should be able to patch resource.")
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
	objectHandler.ServeHTTP(w, inValidDataRequest)
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
	objectHandler.ServeHTTP(w, wrongResourceRequest)
	if w.Code != http.StatusNotFound {
		t.Error("Should return not found.")
	}

	w = httptest.NewRecorder()
	objectHandler.ServeHTTP(w, validRequest)
	if w.Code != http.StatusOK {
		t.Error("Should be able to delete created resource.")
	}
}
