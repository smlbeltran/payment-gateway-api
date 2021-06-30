package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"

	auth_model_req "github.com/smlbeltran/payment-gateway-api/models/authorize/request"
	auth_model_response "github.com/smlbeltran/payment-gateway-api/models/authorize/response"
	capture_model_request "github.com/smlbeltran/payment-gateway-api/models/capture/request"
	capture_model_response "github.com/smlbeltran/payment-gateway-api/models/capture/response"
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

func GetAuthorization(db *bolt.DB, creditInfo *auth_model_req.CreditCard) (*auth_model_response.Authorize, error) {

	authorization_id, _ := authorizeAccount(db, creditInfo)

	var v auth_model_response.Authorize

	// we fetch the newly created authorization response row.
	err := db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Bucket(verifyBucket).Get(authorization_id)
		return json.Unmarshal(value, &v)
	})

	if err != nil {
		panic(err)
	}

	return &v, err
}

func CaptureTransaction(db *bolt.DB, capture capture_model_request.Capture) (*capture_model_response.CaptureResponse, error) {
	var id = capture.AuthorizationId

	// get initial transaction setup for the authorized payment
	captureTransaction, err := getTransactionData(db, id, transactionBucket)

	if err != nil {
		return nil, errors.New("unable to retrieve transaction")
	}

	if captureTransaction["complete"] == true {
		return nil, errors.New("transaction is completed further action not allowed")
	}

	if captureTransaction["refunded"] == true {
		return nil, errors.New("transaction has been refund no further action allowed")
	}

	if captureTransaction["cancelled"] == true {
		return nil, errors.New("transaction has been cancelled no further action allowed")
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
	var response capture_model_response.CaptureResponse
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

func RefundTransaction(db *bolt.DB, reqRefund refund_model_request.Refund) (*refund_model_response.AccountRefundResponse, error) {
	var id = reqRefund.AuthorizationId

	captureTransaction, err := getTransactionData(db, id, transactionBucket)
	if err != nil {
		return nil, errors.New("unable to retrieve capture transaction")
	}

	if captureTransaction["refunded"] == true {
		return nil, errors.New("amount refunded no further action is required")
	}

	if captureTransaction["complete"] == true {
		return nil, errors.New("transaction is completed further action not allowed")
	}

	if captureTransaction["cancelled"] == true {
		return nil, errors.New("transaction has been cancelled no further action allowed")
	}

	refundTransaction, err := getTransactionData(db, id, refundBucket)
	if err != nil {
		return nil, errors.New("unable to retrieve refund transaction")
	}

	captureAmount := int(captureTransaction["captured"].(float64))

	refundTransaction["capture_amount"] = captureAmount
	refundTransaction["refund_amount"] = int(refundTransaction["refund_amount"].(float64)) + reqRefund.Amount

	if refundTransaction["refund_amount"].(int) == captureAmount {
		captureTransaction["refunded"] = true
		t, _ := json.Marshal(captureTransaction)

		err = db.Update(func(tx *bolt.Tx) error {
			err := tx.Bucket(rootBucket).Bucket(transactionBucket).Put([]byte(id), t)
			return err
		})

		if err != nil {
			panic(err)
		}

		r, _ := json.Marshal(refundTransaction)

		err = db.Update(func(tx *bolt.Tx) error {
			value := tx.Bucket(rootBucket).Bucket(refundBucket).Put([]byte(id), r)
			return value
		})

		if err != nil {
			panic(err)
		}
	}
	// TODO:: we need to update the refund amount so that is displayed correctly
	if captureTransaction["refunded"] == false {
		if reqRefund.Amount <= captureAmount {
			if refundTransaction["refund_amount"].(int) > captureAmount {
				return nil, errors.New("value been refunded large than the capture amount allowed")
			}

			r, _ := json.Marshal(refundTransaction)

			err = db.Update(func(tx *bolt.Tx) error {
				value := tx.Bucket(rootBucket).Bucket(refundBucket).Put([]byte(id), r)
				return value
			})

			if err != nil {
				return nil, fmt.Errorf("%s", "unable to update refund transaction")
			}
		} else {
			return nil, errors.New("value greater that the amount allowed to refund")
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

// ===========================Private Functions============================================

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
		var captureAmount int

		captureAmount += receivedAmount

		captureTransaction["captured"] = int(captureTransaction["captured"].(float64)) + captureAmount

		fmt.Println("captured payment:", captureTransaction["captured"])

		if captureTransaction["captured"] == authorizedTransactionAmount {
			captureTransaction["complete"] = true
		}

	} else {
		return nil, errors.New("value greater that the amount authorized to be charged")
	}

	return captureTransaction, nil
}
