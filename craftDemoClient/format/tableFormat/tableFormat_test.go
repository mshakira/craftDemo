package tableFormat

import (
	"testing"
)

func TestExtractContents(t *testing.T) {
	// success case
	type TestStruct struct {
		Name string
		Val int
	}
	sliceData := []TestStruct{{"test", 1234},{"test123", 1234}}

	m := make(map[string]int)
	var header []string
	var contents [][]string

	err := ExtractContents(sliceData, m, &header, &contents)
	if err != nil {
		t.Errorf("Expected nil, got %v\n", err)
	}

	if v, ok := m["Name"]; ok {
		if v != 7 {
			t.Errorf("Expected 4, got %v\n", v)
		}
	} else {
		t.Errorf("Expected `Name` key, but not found")
	}

	if len(header) != 2 {
		t.Errorf("Expected array length of 2, but got %v", len(header))
	}

	if len(contents) != 2 {
		t.Errorf("Expected array length of 1, but got %v", len(contents))
	}

	if contents[0][0] != "test" {
		t.Errorf("Expected string test, but got %v", contents[0][0])
	}

	// failure case 1 - unsupported data type
	type TestFailStruct struct {
		Val float64
	}

	sliceFailData := []TestFailStruct{{1.1}}

	err = ExtractContents(sliceFailData, m, &header, &contents)
	if err == nil {
		t.Errorf("Expected error, got %v\n", err)
	}

	// failure case 2 - unsupported data type
	structData := TestFailStruct{1.1}
	err = ExtractContents(structData, m, &header, &contents)
	if err == nil {
		t.Errorf("Expected error, got %v\n", err)
	}

	// failure case 3 - unsupported data type
	structFailData := []map[string]int{{"Val":1}}
	err = ExtractContents(structFailData, m, &header, &contents)
	if err == nil {
		t.Errorf("Expected error, got %v\n", err)
	}
}

func TestPrintHeader(t *testing.T) {
	m := map[string]int{"Name":4}
	header := []string{"Name"}

	str := PrintHeader(m,&header)
	testStr := "Name   \n#######\n"

	if str != testStr {
		t.Errorf("Expected string %s, but got %s", testStr, str)
	}

	m = map[string]int{"Neque porro quisquam est qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit":92}
	header = []string{"Neque porro quisquam est qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit"}

	str = PrintHeader(m,&header)
	testStr = "Neque porro quisquam est qui dolorem ipsum quia dolor sit am   \n###############################################################\n"

	if str != testStr {
		t.Errorf("Expected string %s, but got %s", testStr, str)
	}

}

func TestPrintContents(t *testing.T) {
	m := map[string]int{"Name":4}
	header := []string{"Name"}

	str := "test"
	contents := [][]string{{str}}

	fmtStr := PrintContents(m,&header,&contents)

	expectedStr := "test   \n"
	if fmtStr != expectedStr {
		t.Errorf("Expected string %s, but got %s", expectedStr,fmtStr)
	}

	str = "Neque porro quisquam est qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit"
	contents = [][]string{{str}}
	fmtStr = PrintContents(m,&header,&contents)

	expectedStr = "Neque porro quisquam est qui dolorem ipsum quia dolor sit am \n"
	if fmtStr != expectedStr {
		t.Errorf("Expected string %s, but got %s", expectedStr,fmtStr)
	}

}
