package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMethodHandler(t *testing.T) {
	testMessage := "test message"
	handler := HTTPMethodHandler{
		GET: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(testMessage))
		}),
	}

	getRequest := httptest.NewRequest("GET", "http://example.com/foo", nil)
	postRequest := httptest.NewRequest("POST", "http://example.com/foo", nil)

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
