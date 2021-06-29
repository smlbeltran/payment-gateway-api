package internal

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
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

func GetAuthorization(db *bolt.DB, creditInfo *auth_model_req.CreditCard) (*auth_model_response.Verfication, error) {

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

func CaptureTransaction(db *bolt.DB, capture account_model_request.Account) (*account_model_response.CaptureResponse, error) {
	var id = capture.AuthorizationId

	// get initial transaction setup for the authorized payment
	captureTransaction, err := getTransactionData(db, id, transactionBucket)

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve transaction: %+v", err)
	}

	if captureTransaction["complete"] == true {
		return nil, fmt.Errorf("%s", "Transaction is completed further action not allowed.")
	}

	if captureTransaction["refund"] == true {
		return nil, fmt.Errorf("%s", "Transaction has been refund no further action allowed")
	}

	if captureTransaction["cancelled"] == true {
		return nil, fmt.Errorf("%s", "Transaction has been cancelled no further action allowed")
	}

	transaction, err := captureAmount(capture.Amount, captureTransaction)

	if err != nil {
		return nil, err
	}

	t, _ := json.Marshal(transaction)
	err = updateTransactionData(db, id, t)

	if err != nil {
		return nil, err
	}

	// return updated transaction
	var response account_model_response.CaptureResponse
	err = db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(transactionBucket).Get([]byte(id))
		return json.Unmarshal(value, &response)
	})

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve updated capture transaction date: %s", err)
	}

	return &response, err
}

func CancelTransaction(db *bolt.DB, authorizationID string) (*void_model_response.VoidResponse, error) {
	var id = authorizationID

	// get capture transaction
	captureTransaction, err := getTransactionData(db, id, transactionBucket)

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve capture transaction: %+v", err)
	}

	captureTransaction["cancelled"] = true

	t, _ := json.Marshal(captureTransaction)

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(transactionBucket).Put([]byte(id), t)
		if err != nil {
			return fmt.Errorf("could not cancel transaction entry: %v", err)
		}

		return nil
	})

	var response void_model_response.VoidResponse
	response.Amount = int(captureTransaction["amount"].(float64))
	response.Currency = captureTransaction["currency"].(string)
	return &response, err
}

// ================================================Refund==========================================================

func RefundTransaction(db *bolt.DB, reqRefund refund_model_request.Refund) (*refund_model_response.AccountRefundResponse, error) {
	var id = reqRefund.AuthorizationId

	// get capture transaction
	captureTransaction, err := getTransactionData(db, id, transactionBucket)

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve capture transaction: %+v", err)
	}

	// get refund
	refundTransaction, err := getTransactionData(db, id, refundBucket)

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve refund transaction: %+v", err)
	}

	captureAmount := int(captureTransaction["capture"].(float64))
	refundTransaction["amount"] = int(refundTransaction["amount"].(float64)) + reqRefund.Amount

	if refundTransaction["amount"].(int) == captureAmount {
		captureTransaction["refunded"] = true
		t, _ := json.Marshal(captureTransaction)

		err = db.Update(func(tx *bolt.Tx) error {
			err := tx.Bucket(rootBucket).Bucket(transactionBucket).Put([]byte(id), t)
			return err
		})

		if err != nil {
			panic(err)
		}
	}

	if captureTransaction["refunded"] == false {
		if reqRefund.Amount <= captureAmount {
			if refundTransaction["amount"].(int) > captureAmount {
				return nil, fmt.Errorf("%s", "value been refunded large than the capture amount allowed")
			}

			r, _ := json.Marshal(refundTransaction)

			//add the new updated refund amount to the db
			err = db.Update(func(tx *bolt.Tx) error {
				value := tx.Bucket(rootBucket).Bucket(refundBucket).Put([]byte(id), r)
				return value
			})

			if err != nil {
				return nil, fmt.Errorf("%s", "unable to update refund transaction")
			}
		} else {
			return nil, fmt.Errorf("%s", "greater that the amount allowed to refund ....")
		}

	}

	var rt refund_model_response.AccountRefundResponse
	err = db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(refundBucket).Get([]byte(id))
		return json.Unmarshal(value, &rt)
	})

	if err != nil {
		panic(err)
	}

	return &rt, nil
}

// ================================================Private Functions========================

func authorizeAccount(db *bolt.DB, creditInfo *auth_model_req.CreditCard) ([]byte, error) {
	authenticationID, data, err := generateAuthentication(creditInfo)
	if err != nil {
		return nil, err
	}

	// Add authentication row
	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(rootBucket).Bucket(verifyBucket).Put([]byte(authenticationID), []byte(data))
		if err != nil {
			return fmt.Errorf("could not insert authorization entry: %v", err)
		}

		return nil
	})

	TransactionsInitializer(db, authenticationID, creditInfo)

	return []byte(authenticationID), err
}

func captureAmount(receivedAmount int, captureTransaction map[string]interface{}) (map[string]interface{}, error) {
	authorizedTransactionAmount := int(captureTransaction["amount"].(float64))

	if receivedAmount <= authorizedTransactionAmount {
		// reducing the intial authorized purchase amount until we have
		// charge the full amount.
		receivedAmount += receivedAmount

		captureTransaction["capture"] = receivedAmount

		if captureTransaction["capture"] == authorizedTransactionAmount {
			captureTransaction["complete"] = true
		}

	} else {
		return nil, fmt.Errorf("%s", "Greater that the amount authorized to be charged.")
	}

	return captureTransaction, nil
}
