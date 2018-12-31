package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var testMessage = "test message"

var testDispatcher = HTTPMethodDispatcher{
	GET: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(testMessage))
	}),
}

var getRequest = httptest.NewRequest("GET", "http://example.com/foo", nil)
var postRequest = httptest.NewRequest("POST", "http://example.com/foo", nil)

func TestMethodDispatcher(t *testing.T) {
	get := testDispatcher.GetMethodHandler(getRequest.Method)
	if testDispatcher.GET == nil {
		t.Error("Dispatcher did not return correct method, returned nil.")
	}
	w := httptest.NewRecorder()
	get.ServeHTTP(w, getRequest)
	if w.Code != http.StatusOK {
		t.Error("Request to supplied method should be ok")
	}
	if w.Body.String() != testMessage {
		t.Error("Method did not call function correctly")
	}

	if testDispatcher.POST != nil {
		t.Error("Dispatcher did not return nil for undefined method.")
	}
}

func TestMethodHandler(t *testing.T) {

	handler := HTTPMethodHandler{
		dispatcher: testDispatcher,
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, getRequest)
	if w.Code != http.StatusOK {
		t.Error("Request to supplied method should be ok")
	}
	if w.Body.String() != testMessage {
		t.Error("Method did not call function correctly")
	}

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, postRequest)
	if w.Code != http.StatusMethodNotAllowed {
		t.Error("Request to unsupplied method should give a 405")
	}
}
