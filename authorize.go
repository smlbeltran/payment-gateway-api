package main

import (
	"encoding/json"
	_ "fmt"
	"io/ioutil"
	"net/http"

	"github.com/boltdb/bolt"
	model_req "github.com/smlbeltran/payment-gateway-api/models/authorize/request"
)

type Authorize struct {
	Db *bolt.DB
}

func (a *Authorize) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			panic(err)
		}
	}()

	var c model_req.CreditCard

	err = json.Unmarshal(body, &c)
	if err != nil {
		panic(err)
	}

	// if strings.TrimSpace(strconv.Itoa(c.CardNumber)) == "4000000000000119" {
	// 	// add error message not authorized....
	// 	// fmt.Errorf("%s", "not authorized to continue ")
	// }

	resp, err := getAccountAuthorization(a.Db, &c)

	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(&resp)
}
