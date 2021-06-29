package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/boltdb/bolt"
	api "github.com/smlbeltran/payment-gateway-api/internal"
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

	cardNumber := strings.TrimSpace(strconv.Itoa(c.CardNumber))
	if len(cardNumber) < 16 {
		http.Error(w, "card number lenght is less than the 16 characters ", http.StatusBadRequest)

		return
	}

	err = goluhn.Validate(cardNumber)

	if err != nil {
		http.Error(w, "invalid card number", http.StatusBadRequest)
		return
	}

	if cardNumber == "4000000000000119" {
		http.Error(w, "card number is not allowed contact customer support", http.StatusUnauthorized)
		return
	}

	resp, err := api.GetAuthorization(a.Db, &c)

	if err != nil {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(&resp)
}
