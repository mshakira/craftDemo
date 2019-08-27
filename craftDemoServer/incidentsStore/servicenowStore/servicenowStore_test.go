package servicenowStore_test

import (
	"craftDemoServer/incidentsStore/servicenowStore"
	"testing"
)

func TestGetList(t *testing.T) {
	//var err error

	// Success case - GetList should return expected json
	var err error
	file := "incidents_test.json"
	snst, err := servicenowStore.Init(file)

	expectedStr := `{"Name":"ServiceNowQuery","Report":[{"number":"INC1234","assigned_to":"","description":"","state":"","priority":"","severity":""}]}`
	js, err := snst.GetList()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	jsStr := string(*js)

	if jsStr != expectedStr {
		t.Errorf("Expected %s, got %s", expectedStr, jsStr)
	}

	// failure case - file not found
	snst.File = "file_not_found.json"
	js, err = snst.GetList()

	if err == nil {
		t.Errorf("Expected error, got %v", err)
	}

	// failure case - no json data
	snst.File = "no_json.json"
	js, err = snst.GetList()

	if err == nil {
		t.Errorf("Expected error, got %v", err)
	}
}




