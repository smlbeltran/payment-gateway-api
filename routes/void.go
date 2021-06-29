package routes

import (
	"encoding/json"
	"fmt"
	_ "fmt"
	"io/ioutil"
	"net/http"

	"github.com/boltdb/bolt"
	api "github.com/smlbeltran/payment-gateway-api/internal"
	model_req "github.com/smlbeltran/payment-gateway-api/models/void/request"
)

type Void struct {
	Db *bolt.DB
}

func (v *Void) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			panic(err)
		}
	}()

	var void model_req.Void

	err = json.Unmarshal(body, &void)
	if err != nil {
		panic(err)
	}

	resp, err := api.CancelTransaction(v.Db, void.AuthorizationId)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusUnauthorized)
		return
	}

	if err != nil {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(&resp)
}
