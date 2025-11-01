package model

type PaymentProvider string

type PaymentInfoRequest struct {
	ID string `json:"id"`
}

type PaymentIntentRequest struct {
	Amount     	int64				`json:"amount"`
	Currency   	string 				`json:"currency"`
	CustomerID 	string				`json:"customer_id"`
	Token	   	string				`json:"token"`
	RememberMe	*bool				`json:"remember_me"`
	Metadata   	map[string]string	`json:"metadata"`
	PaymentMethod string				`json:"payment_method"`
}

type PaymentIntentResponse struct {
	ID           string		`json:"id"`
	Amount       int64		`json:"amount"`
	Currency     string		`json:"currency"`
	Status       string		`json:"status"`
	ClientSecret string		`json:"client_secret"`
}

type PaymentProcessorResponse struct {
	ID           		string
	Amount       		int64
	Currency     		string
	Status       		string
	ClientSecret 		string
	PaymentMethodID		string
	PaymentMethodType	string
	PaymentProvider		string
	PaymentMethodStatus	string
}

type PaymentEvent struct {
	Type    			string		`json:"type"`
	PaymentIntent		string		`json:"payment_intent"`
	PaymentMethod		string		`json:"payment_method"`
	Payload 			interface{}
}

type PaymentMethodResponse struct {
	ClientID    		string		`json:"client_id"`
	PaymentMethodID		string		`json:"payment_method_id"`
	PaymentProvider		string		`json:"payment_provider"`
	PaymentMethodType	string		`json:"payment_method_type"`
	PaymentMethodStatus	string		`json:"payment_method_status"`
}