package main

import (
	"craftDemoServer/incidentsStore/servicenowStore"
	"log"
	"net/http"
)

var snst *servicenowStore.ServicenowStore

func main() {
	// initialize incident store
	// TODO: another level of abstraction?
	snst, _ = servicenowStore.Init("")

	// Add the handler for /api/v1/list/incidents api call
	http.HandleFunc("/api/v1/list/incidents", httpHandler)
	//http.ListenAndServe(":3000", nil)

	// enable SSL
	err := http.ListenAndServeTLS(":443", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func httpHandler(w http.ResponseWriter, r *http.Request) {

	// get the response from serviceNow obj

	js, err := snst.GetList()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the content-type header to json
	w.Header().Set("Content-Type", "application/json")
	w.Write(*js)
}
