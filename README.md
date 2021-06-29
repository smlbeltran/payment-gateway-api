# Payment-Gateway-API

E-Commerce is experiencing exponential growth and merchants who sell their goods or
services online need a way to easily collect money from their customers.

The service gives the ability for the merchant to get approval from the banks and
charge a customer for any given product they purchase, as well it allows the merchant
to cancel a transaction or refund a customer.


### Prerequisites

What things you need to install the software and how to install them

```
https://golang.org/doc/install
```

### Installing

A step by step series of examples that tell you how to get a development env running and
execute all requests.

Download the GO Project

```
https://github.com/smlbeltran/payment-gateway-api.git
```

Then use the following command to build the binary, this will be stored in the root of your project

```
go build
```

Once the binary is created run

```
./payment-gateway-api
```

This will set the server for incoming request

## Running the tests

Do to time constrains and work there are no automated tests for this system
at this moment but you can use the following endpoints to test the system.

## Considerations
1. If work would not have interfered I would most likely set some validation in the middleware
   setting my controllers to have less responsability or visibility for certain actions.
2. would have further decoupled some of the internal logic to provide better readability.

## Problem found
I believe some of the test edge cases are not fully correct for the requirements of the
challenge and the behaviour is intended.

e.g
* 4000 0000 0000 0259: capture failure
* 4000 0000 0000 3238: refund failure

1. These tests can't be executed as we are unable to get the credit card number from the
   requests. We are only able to send the `authorization_id` for these.
2. My understanding is the credit card is only validated when running the `/authorize` endpoint where
   we are able to provide the card number within the payload. Once validated and correct we can go down the line
   we other actions.

### Authorization Create
It will return a unique authorization ID that will be used in all next API calls.

Endpoint

```
POST curl -XPOST http://localhost:8000/authorize 
```

Request
```
{
    "card_number":4000000000000259,
    "card_expiry_month":1,
    "card_expiry_year":21, 
    "cvv": 222,
    "amount": 10,
    "currency": "GBP"
}
```

Response
```
{
    "authorization_id":"xxxxxx",
    "status":1,
    "amount":10,
    "currency":"GBP"
}
```

Full Example within the Terminal
```
curl -XPOST http://localhost:8000/authorize -d '{"card_number":4000000000000259, "card_expiry_month":1, "card_expiry_year":21, "cvv": 222, "amount": 10, "currency": "GBP"}' -H "Content-Type: application/json"
```


### Capture Amount
Will capture the money on the customer bank. It can be called
multiple times with an amount that should not be superior to the amount authorised in
the first call. 

For example, a 10£ authorisation can be captured 2 times with a 4£ and 6£ order.

Endpoint

```
POST curl -XPOST http://localhost:8000/capture
```

Request

```
{
    "authorization_id":"xxxxxx",
    "amount":6
}
```

Response
```
{
    "amount":10,
    "captured":6,
    "currency":"GBP"
}
```

Full Example within the Terminal
```
curl -XPOST http://localhost:8000/capture -d '{"authorization_id":"xxxxxx", "amount":6}' -H "Content-Type: application/json"
```

### Cancel Transaction

This will cancel the whole transaction without billing the customer.

Endpoint

```
POST http://localhost:8000/void
```

Request

```
{
   "authorization_id":"xxxxxx",
}
```

Response
```
{
   "status":0,
   "amount":10,
   "currency":"GBP"
}
```

Full Example within the Terminal
```
curl -XPOST http://localhost:8000/void -d '{"authorization_id":"xxxxxx"}' -H "Content-Type: application/json"
```

### Refund Account

Will refund the money taken from the customer bank account.

Endpoint
```
POST http://localhost:8000/refund
```

Request
```
{
   "authorization_id":"xxxxxx",
   "amount":3
}
```

Response
```
{
    "capture_amount":6,
    "refund_amount":3,
    "currency":"GBP"
}
```

Full Example within the Terminal
```
curl -XPOST http://localhost:8000/refund -d '{"authorization_id":"xxxxxx", "amount":3}' -H "Content-Type: application/json"
```

### View Transaction Status Details

```
Will allow to view the state of the current transaction.
below are the states which the transaction can be set to according to the behaviour:

* if the full amount of the authorized request has been captured then the status `complete` will be set to `true`
   and no further actions can be taken

* if a full refund is made then the status `refunded` will be set to `true`
  and no further actions can be taken

* if a transaction has been cancelled then the status `cancelled` will be set to `true` 
  and no further actions can be taken
```

Endpoint
```
GET http://localhost:8000/transaction/{authorizationID}
```

Response
```
{
    "amount":10,
    "cancelled":false,
    "captured":6,
    "complete":false,
    "currency":"GBP",
    "refunded":false
}
```

## Built With

* [Golang](https://golang.org/) - The Go Programming Language.
* [MUX](https://github.com/gorilla/mux) - A powerful URL router and dispatcher for golang.
* [BoltDb](https://github.com/boltdb/bolt) - An embedded key/value database for Go.
* [goluhn](https://github.com/ShiraazMoollatjie/goluhn) - Luhn credit check validator