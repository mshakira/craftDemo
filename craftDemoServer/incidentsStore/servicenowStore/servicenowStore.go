package servicenowStore

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// The store obj has File as we get the data from a file
type ServicenowStore struct {
	File string
}

// Entire Incidents object
type Incidents struct {
	Name    string `json:"Name"`
	Report []Incident `json:"Report"`
}

// Individual incident object
type Incident struct {
	Number string `json:"number"`
	AssignedTo string `json:"assigned_to"`
	Description string `json:"description"`
	State string `json:"state"`
	Priority string `json:"priority"`
	Severity string `json:"severity"`
}

/* 
Initializes the servicenow object with fileStore to read
 */
func Init(file string) (*ServicenowStore, error){
	var snst ServicenowStore
	// for testing purpose, send file name
	if file != "" {
		snst.File = file
	} else { // use default file
		snst.File = "incidents.json"
	}
	// initialize session with serviceNow store
	return &snst, nil
}

/*
Add GetList method to ServicenowStore
Read the file, extract json objects and map it to required struct
Return final json
 */
func (snst *ServicenowStore) GetList() (*[]byte, error)  {
	jsonFile, err := os.Open(snst.File)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var incidents Incidents

	// encode bytes to incidents struct object by mapping required fields
	err = json.Unmarshal(byteValue, &incidents)

	if err != nil {
		return nil, err
	}

	// decode the json to string
	js, err := json.Marshal(&incidents)

	if err != nil {
		return nil, err
	}

	return &js, nil
}