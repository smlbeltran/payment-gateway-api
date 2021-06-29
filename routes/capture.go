package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/boltdb/bolt"
	api "github.com/smlbeltran/payment-gateway-api/internal"
	model_req "github.com/smlbeltran/payment-gateway-api/models/capture/request"
)

type Capture struct {
	Db *bolt.DB
}

func (c *Capture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			panic(err)
		}
	}()

	var account model_req.Account

	err = json.Unmarshal(body, &account)
	if err != nil {
		panic(err)
	}

	resp, err := api.CaptureTransaction(c.Db, account)

	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(&resp)
}
