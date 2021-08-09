package externals

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"go-worker/models"
	"go-worker/utils"
)

var responseBody = `{"call_id": "e21b0dda-6566-402a-8f8c-0657e5b87eeb", "charge_amount": 1.2}`

// TestPositiveBillUser - test a successful api response
func TestPositiveBillUser(t *testing.T) {

	check := assert.New(t)

	// disable transport swap for http mock
	gorequest.DisableTransportSwap = true

	// create request handler for mocking
	requestHandler := utils.NewRequestHandler("balance_api")
	httpmock.ActivateNonDefault(requestHandler.Handler.Client)
	defer httpmock.DeactivateAndReset()

	// prepare bill event
	billEvent := getBillingEvent()

	// mock http request
	httpmock.RegisterResponder("POST", "https://balance-svc-dev.com/Billing/1",
		httpmock.NewStringResponder(200, responseBody))

	// call balance api
	balanceRequestHandler := getBalanceHandler(requestHandler)
	response, isSuccessful := balanceRequestHandler.BillUser(billEvent)
	if isSuccessful {
		check.Equal(responseBody, string(response))
	}

	// get the amount of calls for the registered responder
	info := httpmock.GetCallCountInfo()
	check.Equal(1, info["POST https://balance-svc-dev.com/Billing/1"])
}

// TestNegativeBillUser - test a failure api response
func TestNegativeBillUser(t *testing.T) {

	check := assert.New(t)

	// disable transport swap for http mock
	gorequest.DisableTransportSwap = true

	// create request handler for mocking
	requestHandler := utils.NewRequestHandler("balance_api")
	httpmock.ActivateNonDefault(requestHandler.Handler.Client)
	defer httpmock.DeactivateAndReset()

	// prepare bill event
	billEvent := getBillingEvent()

	// mock http request
	httpmock.RegisterResponder("POST", "https://balance-svc-dev.com/Billing/1",
		httpmock.NewStringResponder(400, responseBody))

	// call balance api
	balanceRequestHandler := getBalanceHandler(requestHandler)
	_, isSuccessful := balanceRequestHandler.BillUser(billEvent)
	check.Equal(false, isSuccessful)

	// get the amount of calls for the registered responder
	info := httpmock.GetCallCountInfo()
	check.Equal(1, info["POST https://balance-svc-dev.com/Billing/1"])
}

// getBillingEvent - prepares a bill event
func getBillingEvent() (billEvent models.BillingEvent) {
	billEvent.CallID = "e21b0dda-6566-402a-8f8c-0657e5b87eeb"
	billEvent.UserID = 1
	billEvent.ProductID = 2
	billEvent.AnswerTime = "2021-07-01 00:30:00"
	billEvent.HangupTime = "2021-07-01 01:00:00"
	return
}

// getBalanceHandler - handler for calling balance api
func getBalanceHandler(requestHandler *utils.RequestHandler) (balanceRequestHandler BalanceRequestHandler) {
	balanceRequestHandler = BalanceRequestHandler{
		URL:        "https://balance-svc-dev.com",
		Username:   "test",
		Password:   "test",
		Timeout:    1,
		RetryCount: 2,
		Request:    requestHandler,
		Log:        logrus.New().WithField("prefix", fmt.Sprintf("WorkerID:%d", 1)),
	}
	return
}
