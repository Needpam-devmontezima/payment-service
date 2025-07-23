package model

import (
	"time"
)

type Status string

type Transaction struct {
	ID					string
	InternalReference 	string
	Amount				int64
	Currency        	string
	PaymentIntentID  	string
	TxStatus	        string
	CustomerID      	string
	PaymentMethodID		string
	PaymentProvider 	string
	PaymentMethodType 	string
	PaymentMethodStatus	string
	SavePaymentMethod	*bool
	CreatedAt			time.Time
	UpdatedAt			time.Time
	Metadata			map[string]string       
}