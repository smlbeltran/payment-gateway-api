package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	db "github.com/smlbeltran/payment-gateway-api/internal"
	routes "github.com/smlbeltran/payment-gateway-api/routes"
)

func main() {

	router := mux.NewRouter()

	//initialize db session
	DB, err := db.SetupDB()
	if err != nil {
		panic(err)
	}

	router.Handle("/authorize", &routes.Authorize{Db: DB}).Methods("POST")
	router.Handle("/capture", &routes.Capture{Db: DB}).Methods("POST")
	router.Handle("/void", &routes.Void{Db: DB}).Methods("POST")
	router.Handle("/refund", &routes.Refund{Db: DB}).Methods("POST")

	router.Handle("/transaction/{applicationID}", &routes.Transaction{Db: DB}).Methods("GET")

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8001",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

}
