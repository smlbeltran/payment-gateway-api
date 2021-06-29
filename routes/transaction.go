package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

var transactionBucket = []byte("TRANSACTION_CAPTURE")
var rootBucket = []byte("DB")

type Transaction struct {
	Db *bolt.DB
}

func (t *Transaction) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	transactionID := mux.Vars(r)["applicationID"]

	var transaction map[string]interface{}

	err := t.Db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(transactionBucket).Get([]byte(string(transactionID)))
		return json.Unmarshal(value, &transaction)
	})

	if err != nil {
		fmt.Println("something went worng here")
	}

	json.NewEncoder(w).Encode(transaction)
}
