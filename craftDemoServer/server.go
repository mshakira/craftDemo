package main

import (
	"craftDemoServer/incidentsStore/servicenowStore"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"time"
)

var snst *servicenowStore.ServicenowStore

func main() {

	mux := http.NewServeMux()

	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)

	// initialize incident store
	// TODO: another level of abstraction?
	snst, _ = servicenowStore.Init("")
	log.Info("Server starting...")
	log.Info("Initializing serviceNow store")

	// Add the handler for /api/v1/list/incidents api call
	mux.HandleFunc("/api/v1/list/incidents", httpHandler)
	//http.ListenAndServe(":3000", nil)

	// enable SSL
	err := http.ListenAndServeTLS(":443", "server.crt", "server.key", RequestLogger(mux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func RequestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		targetMux.ServeHTTP(w, r)

		// log request by who(IP address)
		requesterIP := r.RemoteAddr

		log.Printf(
			"%s\t\t%s\t\t%s\t\t%v",
			r.Method,
			r.RequestURI,
			requesterIP,
			time.Since(start),
		)
	})
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
