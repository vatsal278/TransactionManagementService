# TransactionManagementService

[![Build](https://github.com/vatsal278/TransactionManagementService/actions/workflows/build.yml/badge.svg)](https://github.com/vatsal278/TransactionManagementService/actions/workflows/build.yml) [![Test Cases](https://github.com/vatsal278/TransactionManagementService/actions/workflows/test.yml/badge.svg)](https://github.com/vatsal278/TransactionManagementService/actions/workflows/test.yml) [![Codecov](https://codecov.io/gh/vatsal278/TransactionManagementService/branch/main/graph/badge.svg)](https://codecov.io/gh/vatsal278/TransactionManagementService)

* This service was created using Golang.
* This service has used clean code principle and appropriate go directory structure.
* This service is completely unit tested and all the errors have been handled.

## Starting the Transaction Management service

* Start the Docker container for mysql with command :
```
docker run --publish 9075:3306 -d mysql
```

* Start the Api locally with command :
```
go run .\cmd\TransactionManagementService\main.go
```
### You can test the api using post man, just import the [Postman Collection](./docs/transactionService.postman_collection.json) into your postman app.
### To check the code coverage
```
cd docs
go tool cover -html=coverage
```
## Transaction Management service:

This application is split up into multiple components, each having a particular feature and use case. This will allow individual scale up/down and can be started up as micro-services.

HTTP calls are made across micro-services.

*For testing individual services, these can be via direct HTTP calls*


All requests & responses follow json encoding.
Requests are specific to the concerned endpoint while responses are of the following json format & specification:
>
>    Response Header: HTTP code
>
>    Response Body(json):
>    ```json
>    {
>       "status": <HTTP status code>,
>       "message": "<message>",
>       "data": {
>        // object to contain the appropriate response data per API
>       }
>    }
>    ```

## Transaction Management Service Endpoints

## List Transactions
A user hits this endpoint in order to view their transactions. It uses pagenation to list a specefic number of records at once.For viewing a specefic transaction we can specify the transaction id of that record.
There will be jwt token containing userid in cookie for authentication of user.
#### Specification:
Method: `GET`

Path (list of transaction): `/transactions`

query parameters:
- `page` : specefic page which user wants to use
- `limit` : total number of records in that page

Request Body: `not required.`

Success to follow response as specified:

Response Header: HTTP 200

Response Body(json):
```json
{
  "status": 200,
  "message": "SUCCESS",
  "data": [{
    "account_number": <account number as int>,
    "transaction_id":"<full transaction id> as string",
    "amount": <amount of transaction>as float,
    "transfer_to":<account number of receiver as int>,
    "created_at": "date of transaction" DD-MM-YYY format,
    "updated_at": "updated date of transaction" DD-MM-YYY format,
    "status": <status of transaction ,approved or rejected as string>,
    "type" :<credit Or debit type of transaction as string>,
    "comment":<comment about the transaction as string>
  }]
}
```

## Do Transaction
This endpoint is used to do a new transaction. It is a post endpoint which is used to update the database with latest transaction and its details.
This endpoint stores the transaction data along with the user_id which can be obtained from cookie.Once insertion is successful transaction details are sent to account management service for updating income and spends.
#### Specification:
Method: `POST`

Path: `/transactions`

Request Body:
```json
{
  "account_number":<account number an int>,
  "amount": total amount of the transaction as float,
  "status":"approved or rejected as string",
  "transafer_to":<account_number as int>,
  "comment":"comment if any as string",
  "type":"debit or credit as string"
}
```

Success to follow response as specified:

Response Header: HTTP 200

Response Body(json):
```json
{
  "status": 201,
  "message": "SUCCESS",
  "data": nil
}
```

## Download Transaction Details
This endpoint is used to download the transaction detail of a specific transaction as a pdf. It fetches user details by making call to user management service
#### Specification:
Method: `GET`

Path: `transactions/download/{transaction_id}`

Request Body: `nil`

Url Parameter: `transaction_id of transaction that needs to be downloaded in url path.`

Success to follow response as specified:

Response Header: HTTP 200

Response Body(pdf):Pdf file will get downloaded

## AccManagementSvc Middlewares

1. ExtractUser: extracts the user_id from the cookie passed in the request and forwards it in the context for downstream processing.
2. Caching middleware