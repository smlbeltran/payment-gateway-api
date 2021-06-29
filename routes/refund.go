package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/boltdb/bolt"
	api "github.com/smlbeltran/payment-gateway-api/internal"
	model_req "github.com/smlbeltran/payment-gateway-api/models/refund/request"
)

type Refund struct {
	Db *bolt.DB
}

func (c *Refund) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			panic(err)
		}
	}()

	var refund model_req.Refund

	err = json.Unmarshal(body, &refund)
	if err != nil {
		panic(err)
	}

	resp, err := api.RefundTransaction(c.Db, refund)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusUnauthorized)
		return
	}

	if err != nil {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(&resp)
}
