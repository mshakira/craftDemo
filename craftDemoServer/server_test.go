package main

import (
	"craftDemoServer/incidentsStore/servicenowStore"
	"net/http"
	"net/http/httptest"
	"testing"
)


func TestHttpHandler (t *testing.T) {
	//
	snst, _ = servicenowStore.Init("no_file.json")
	req, err := http.NewRequest("GET", "/api/v1/list/incidents", nil)
	if err != nil {
		t.Fatal(err)
	}
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httpHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status == http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want non 200",
			status)
	}

	snst, _ = servicenowStore.Init("")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	/*// Check the response body is what we expect.
	expected := `{"alive": true}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}*/
}

/*func TestIncidentServer_Run(t *testing.T) {
	incidentServer := IncidentServer{}
	incidentServer.snStore, _ = servicenowStore.Init("")
}*/
