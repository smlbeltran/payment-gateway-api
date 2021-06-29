package internal

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	auth_model_req "github.com/smlbeltran/payment-gateway-api/models/authorize/request"
	auth_model_response "github.com/smlbeltran/payment-gateway-api/models/authorize/response"
	"math/rand"
)

func getTransactionData(db *bolt.DB, id string, bucket []byte) (map[string]interface{}, error) {
	var transaction map[string]interface{}

	err := db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(bucket).Get([]byte(id))
		return json.Unmarshal(value, &transaction)
	})

	return transaction, err
}

func updateTransactionData(db *bolt.DB, id string, t []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(transactionBucket).Put([]byte(id), t)
		if err != nil {
			return fmt.Errorf("could not update transaction entry: %v", err)
		}

		return nil
	})
}

func generateAuthentication(creditInfo *auth_model_req.CreditCard) (string, []byte, error) {
	data := &auth_model_response.Verfication{
		AuthorizationId: fmt.Sprintf("%v", rand.Intn(4666778181156224-4666000000000000)),
		Status:          1,
		Amount:          creditInfo.Amount,
		Currency:        creditInfo.Currency,
	}

	var key = fmt.Sprintf("%v", data.AuthorizationId)
	var value, err = json.Marshal(data)

	if err != nil {
		return "", nil, fmt.Errorf("could not generate authentication data entry: %v", err)
	}

	return key, value, nil
}

func TransactionsInitializer(db *bolt.DB, authorizationID string, creditInfo *auth_model_req.CreditCard) error {
	captureTransactionSetup, err := json.Marshal(map[string]interface{}{
		"amount":    creditInfo.Amount,
		"currency":  creditInfo.Currency,
		"captured":  0,
		"complete":  false,
		"refunded":  false,
		"cancelled": false,
	})

	if err != nil {
		panic(err)
	}

	refundTransactionSetup, err := json.Marshal(map[string]interface{}{
		"amount":   0,
		"currency": creditInfo.Currency,
	})

	if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(transactionBucket).Put([]byte(authorizationID), []byte(captureTransactionSetup))
		if err != nil {
			return fmt.Errorf("could not insert capture transaction entry: %v", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(refundBucket).Put([]byte(authorizationID), []byte(refundTransactionSetup))
		if err != nil {
			return fmt.Errorf("could not insert refund transaction entry: %v", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
