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
	void_model_response "github.com/smlbeltran/payment-gateway-api/models/void/response"
)

var rootBucket = []byte("DB")
var verifyBucket = []byte("VERIFY")
var transactionBucket = []byte("TRANSACTION")

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
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}

	fmt.Println("DB Setup Done")
	return db, nil
}

// ================================================Authorization==========================================================

func getAccountAuthorization(db *bolt.DB, creditInfo *auth_model_req.CreditCard) (*auth_model_response.Verfication, error) {

	id, _ := generateAuthorization(db, creditInfo)

	var verifacation auth_model_response.Verfication

	err := db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(verifyBucket).Get(id)
		return json.Unmarshal(value, &verifacation)
	})

	if err != nil {
		panic(err)
	}
	// GET DETAILS BEFORE UPDATE
	detailsTransaction, err := json.Marshal(map[string]interface{}{
		"amount":   verifacation.Amount,
		"currency": verifacation.Currency,
	})

	if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(transactionBucket).Put([]byte(verifacation.VerificationId), []byte(detailsTransaction))
		if err != nil {
			return fmt.Errorf("could not insert transaction entry: %v", err)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return &verifacation, err
}

func generateAuthorization(db *bolt.DB, creditInfo *auth_model_req.CreditCard) ([]byte, error) {
	auth := &auth_model_response.Verfication{
		VerificationId: fmt.Sprintf("%v", rand.Intn(4666778181156223-4666000000000000)),
		Status:         1,
		Amount:         creditInfo.Amount,
		Currency:       creditInfo.Currency,
	}

	var key = fmt.Sprintf("%v", auth.VerificationId)
	var value, err = json.Marshal(auth)

	if err != nil {
		return nil, fmt.Errorf("could not json marshal entry: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(verifyBucket).Put([]byte(key), []byte(value))
		if err != nil {
			return fmt.Errorf("could not insert authorization entry: %v", err)
		}

		return nil
	})

	return []byte(key), err
}

// ================================================Void==========================================================

func cancelTransaction(db *bolt.DB, authorizationID string) (*void_model_response.VoidResponse, error) {
	var id = authorizationID

	//GET DETAILS BEFORE DELETING
	var v void_model_response.VoidResponse

	err := db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(transactionBucket).Get([]byte(id))
		return json.Unmarshal(value, &v)
	})

	if err != nil {
		panic(err)
	}

	// DELELTE ENTRY
	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(transactionBucket).Delete([]byte(id))
		if err != nil {
			return fmt.Errorf("could not delete entry: %v", err)
		}

		return nil
	})

	return &v, err
}

// ================================================Capture==========================================================
// need to fix reduce amount.........
func billAccount(db *bolt.DB, account account_model_request.Account) (*account_model_response.AccountBillingResponse, error) {
	var id = account.TransactionId

	// GET DETAILS BEFORE UPDATE
	// GET remaing info (amount left to transac)
	var accountBilling account_model_response.AccountBillingResponse

	err := db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(transactionBucket).Get([]byte(id))
		return json.Unmarshal(value, &accountBilling)
	})

	if err != nil {
		panic(err)
	}

	// NOTE:: need to fix the amount issue.......
	transactionAmount := int(accountBilling.Amount)

	fmt.Println("from db:", transactionAmount)
	fmt.Println("from request account:", account.Amount)

	if account.Amount < transactionAmount {
		accountBilling.Amount = transactionAmount - account.Amount
	} else {
		panic("greater that the amount authorized")
	}

	transaction, _ := json.Marshal(accountBilling)

	// ADD ENTRY INTO TRANSACTION
	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(transactionBucket).Put([]byte(id), transaction)
		if err != nil {
			return fmt.Errorf("could not delete entry: %v", err)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return &accountBilling, err
}
