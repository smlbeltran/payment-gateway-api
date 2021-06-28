package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter()

	//initialize db session
	db, err := SetupDB()
	if err != nil {
		panic(err)
	}

	router.Handle("/authorize", &Authorize{db}).Methods("POST")
	router.Handle("/void", &Void{db}).Methods("POST")
	router.Handle("/capture", &Capture{db}).Methods("POST")
	router.Handle("/refund", &Refund{db}).Methods("POST")

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8001",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

}
