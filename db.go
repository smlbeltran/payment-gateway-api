package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"math/rand"
	"time"

	auth_model_req "github.com/smlbeltran/payment-gateway-api/models/authorize/request"
	auth_model_response "github.com/smlbeltran/payment-gateway-api/models/authorize/response"
	account_model_request "github.com/smlbeltran/payment-gateway-api/models/capture/request"
	account_model_response "github.com/smlbeltran/payment-gateway-api/models/capture/response"
	refund_model_request "github.com/smlbeltran/payment-gateway-api/models/refund/request"
	refund_model_response "github.com/smlbeltran/payment-gateway-api/models/refund/response"
	void_model_response "github.com/smlbeltran/payment-gateway-api/models/void/response"
)

var rootBucket = []byte("DB")
var verifyBucket = []byte("VERIFY")
var transactionBucket = []byte("TRANSACTION_CAPTURE")
var refundBucket = []byte("TRANSACTION_REFUND")

func SetupDB() (*bolt.DB, error) {
	db, err := bolt.Open("main.db", 0600, &bolt.Options{Timeout: 1 * time.Second})

	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte(rootBucket))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte(verifyBucket))
		if err != nil {
			return fmt.Errorf("could not create verify bucket: %v", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte(transactionBucket))
		if err != nil {
			return fmt.Errorf("could not create transaction bucket: %v", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte(refundBucket))
		if err != nil {
			return fmt.Errorf("could not create refund bucket: %v", err)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}

	fmt.Println("DB Setup Done")
	return db, nil
}

// ================================================Authorization==========================================================

func getAuthorization(db *bolt.DB, creditInfo *auth_model_req.CreditCard) (*auth_model_response.Verfication, error) {

	authorization_id, _ := authorizeAccount(db, creditInfo)

	var v auth_model_response.Verfication

	// we fetch the newly created verification response row.
	err := db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(verifyBucket).Get(authorization_id)
		return json.Unmarshal(value, &v)
	})

	if err != nil {
		panic(err)
	}

	return &v, err
}

// ================================================Void==========================================================

func cancelTransaction(db *bolt.DB, authorizationID string) (*void_model_response.VoidResponse, error) {
	var id = authorizationID

	var transaction map[string]interface{}

	//GET DETAILS BEFORE DELETING
	var v void_model_response.VoidResponse

	err := db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(transactionBucket).Get([]byte(id))
		return json.Unmarshal(value, &transaction)
	})

	if err != nil {
		panic(err)
	}

	transaction["cancelled"] = true

	t, _ := json.Marshal(transaction)

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(transactionBucket).Put([]byte(id), t)
		if err != nil {
			return fmt.Errorf("could not cancel transaction entry: %v", err)
		}

		return nil
	})

	return &v, err
}

// ================================================Capture==========================================================
func billAccount(db *bolt.DB, account account_model_request.Account) (*account_model_response.AccountBillingResponse, error) {
	var transaction map[string]interface{}

	var id = account.AuthorizationId

	// get initial transaction setup for the authorized payment
	err := db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(transactionBucket).Get([]byte(id))
		return json.Unmarshal(value, &transaction)
	})

	if err != nil {
		panic(err)
	}

	if transaction["cancelled"] == true {
		panic("this transaction has been cancelled no further operations are allowed")
	}

	if transaction["complete"] == true {
		panic("you can't charge more as this is transaction is completed.")
	}

	remainingAmount := int(transaction["amount"].(float64))

	if account.Amount <= remainingAmount {
		transaction["amount"] = remainingAmount - account.Amount

		if transaction["amount"] == 0 {
			transaction["complete"] = true
		}

	} else {
		panic("greater that the amount authorized to be charged....")
	}

	t, _ := json.Marshal(transaction)

	// update transaction
	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(transactionBucket).Put([]byte(id), t)
		if err != nil {
			return fmt.Errorf("could not update transaction entry: %v", err)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	// return updated transaction
	var response account_model_response.AccountBillingResponse
	err = db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(transactionBucket).Get([]byte(id))
		return json.Unmarshal(value, &response)
	})

	if err != nil {
		panic(err)
	}

	return &response, err
}

// ================================================Refund==========================================================

func refundAccount(db *bolt.DB, refund refund_model_request.Refund) (*refund_model_response.AccountRefundResponse, error) {
	return nil, nil
}

// ================================================Util Functions==========================================================

func authorizeAccount(db *bolt.DB, creditInfo *auth_model_req.CreditCard) ([]byte, error) {
	randomVerification := &auth_model_response.Verfication{
		AuthorizationId: fmt.Sprintf("%v", rand.Intn(4666778181156224-4666000000000000)),
		Status:          1,
		Amount:          creditInfo.Amount,
		Currency:        creditInfo.Currency,
	}

	var key = fmt.Sprintf("%v", randomVerification.AuthorizationId)
	var value, err = json.Marshal(randomVerification)

	if err != nil {
		return nil, fmt.Errorf("could not json marshal entry: %v", err)
	}

	// Add to database our random verification card response
	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(verifyBucket).Put([]byte(key), []byte(value))
		if err != nil {
			return fmt.Errorf("could not insert authorization entry: %v", err)
		}

		return nil
	})

	setChargableTransaction(db, key, creditInfo)

	return []byte(key), err
}

func setChargableTransaction(db *bolt.DB, authorizationID string, creditInfo *auth_model_req.CreditCard) {

	transaction, err := json.Marshal(map[string]interface{}{
		"amount":    creditInfo.Amount,
		"currency":  creditInfo.Currency,
		"complete":  false,
		"refunded":  false,
		"cancelled": false,
	})

	if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(transactionBucket).Put([]byte(authorizationID), []byte(transaction))
		if err != nil {
			return fmt.Errorf("could not insert transaction entry: %v", err)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
}
