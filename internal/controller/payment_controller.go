package controller

import (
	"log"
	"io"
	"time"
	"context"
	"net/http"
	"encoding/json"
	"github.com/danielmoisemontezima/zw-payment-service/pkg/utils"
	"github.com/danielmoisemontezima/zw-payment-service/internal/model"
	"github.com/danielmoisemontezima/zw-payment-service/internal/service"
)

type PaymentController struct {
	service *service.PaymentService
}

func NewPaymentController(service *service.PaymentService) *PaymentController {
	return &PaymentController{service: service}
}

func (c *PaymentController) CreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	var provider model.PaymentProvider = model.PaymentProvider(r.PathValue("provider"))

	// Create base context from the HTTP request
    ctx := r.Context()
    
    // Add payment-service-specific timeout (5 seconds)
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

	var req model.PaymentIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Set default false if remember_me wasn't specified in JSON
	if req.RememberMe == nil {
		defaultRememberMe := false
		req.RememberMe = &defaultRememberMe
	}

	response, err := c.service.CreatePaymentIntent(ctx, provider, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (c *PaymentController) ChargeClient(w http.ResponseWriter, r *http.Request) {
	var provider model.PaymentProvider = model.PaymentProvider(r.PathValue("provider"))

	// Create base context from the HTTP request
    ctx := r.Context()
    
    // Add payment-specific timeout (5 seconds)
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

	var req model.PaymentIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	response, err := c.service.ChargeClient(ctx, provider, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (c *PaymentController) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var provider model.PaymentProvider = model.PaymentProvider(r.PathValue("provider"))

	// Create base context from the HTTP request
    ctx := r.Context()
    
    // Add payment-specific timeout (5 seconds)
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

	var req model.PaymentInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	response, err := c.service.GetPaymentIntent(ctx, provider, req.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, response)	
}

func (c *PaymentController) ParseWebhook(w http.ResponseWriter, r *http.Request) {
	var provider model.PaymentProvider = model.PaymentProvider(r.PathValue("provider"))

	// Create base context from the HTTP request
    ctx := r.Context()
    
    // Add payment-specific timeout (5 seconds)
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid body")
		return
	}

	event, err := c.service.ParseWebhook(ctx, provider, rawBody, r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("webhook--Event: %+v", event)
	utils.RespondWithJSON(w, http.StatusOK, event)
}

func (c *PaymentController) GetUserPMethods(w http.ResponseWriter, r *http.Request) {
	var userId string = r.PathValue("id")

	// Create base context from the HTTP request
    ctx := r.Context()
    
    // Add payment-service-specific timeout (5 seconds)
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

	response, err := c.service.GetUserPMethods(ctx, userId)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (c *PaymentController) GetHealthCheck(w http.ResponseWriter, r *http.Request) {
    // Create the health check response
    response := map[string]string{
        "status": "OK",
    }
    
    // Send JSON response with 200 status
    utils.RespondWithJSON(w, http.StatusOK, response)
}
