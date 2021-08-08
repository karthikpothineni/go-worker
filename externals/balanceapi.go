package externals

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"go-worker/config"
	"go-worker/models"
	"go-worker/utils"
)

// BalanceRequestHandler - holds customizable fields to request Balance API
type BalanceRequestHandler struct {
	URL        string
	Username   string
	Password   string
	Timeout    int
	RetryCount int
	Request    *utils.RequestHandler
	Log        *logrus.Entry
}

// NewBalanceRequestHandler - returns a new object for BalanceRequestHandler
func NewBalanceRequestHandler(log *logrus.Entry) *BalanceRequestHandler {
	cfg := config.GetConfig()
	balanceHandler := &BalanceRequestHandler{
		URL:        cfg.GetString("balance_service.url"),
		Username:   cfg.GetString("balance_service.username"),
		Password:   cfg.GetString("balance_service.password"),
		Timeout:    cfg.GetInt("balance_service.timeout"),
		RetryCount: cfg.GetInt("balance_service.retry_count"),
		Request:    utils.NewRequestHandler("balance_api"),
		Log:        log,
	}
	return balanceHandler
}

// BillUser - bills the user based on call duration
func (br *BalanceRequestHandler) BillUser(billEvent models.BillingEvent) ([]byte, bool) {

	// prepare url and body
	path := fmt.Sprintf("Billing/%s", billEvent.UserID)
	data := make(map[string]interface{})
	data["product_id"] = billEvent.ProductID
	data["call_id"] = billEvent.CallID
	data["answer_time"] = billEvent.AnswerTime
	data["hangup_time"] = billEvent.HangupTime

	// make api request
	statusCode, response := br.makeRequest(http.MethodPost, path, data)
	if statusCode != http.StatusOK {
		br.Log.Errorf("Failed to call billing api: %d -- %s", statusCode, string(response))
		return response, false
	}
	br.Log.WithFields(logrus.Fields{"response": string(response), "response_code": statusCode}).Debug("Response for billing from API")
	return response, true
}

// makeRequest - prepares request and makes an API call
func (br *BalanceRequestHandler) makeRequest(requestMethod, path string, params map[string]interface{}) (int, []byte) {
	code, response, _ := br.Request.Fetch(&utils.RequestSpecifications{
		URL:        fmt.Sprintf("%v/%v", br.URL, path),
		Params:     params,
		UseAuth:    true,
		Username:   br.Username,
		Password:   br.Password,
		Timeout:    br.Timeout,
		RetryCount: br.RetryCount,
		HTTPMethod: requestMethod,
		Log:        br.Log,
	})
	return code, response
}
