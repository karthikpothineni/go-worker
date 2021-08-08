package utils

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

const (
	// defaultRetryInterval specified in seconds
	defaultRetryInterval = 1
	// default maximum idle connection
	defaultMaxIdleConnection = 100
	// default keep-alive duration
	defaultKeepAliveTime = 30 * time.Second
	// default idle connection timeout
	defaultIdleConnectionTimeout = 30 * time.Second
	// default timeout for the default transport in seconds
	defaultTimeout = 10
	// json request type
	jsonRequestType = "json"
)

var (
	DefaultTransport = &http.Transport{
		DialContext: (&net.Dialer{
			KeepAlive: defaultKeepAliveTime,
		}).DialContext,
		MaxIdleConns:        defaultMaxIdleConnection,
		IdleConnTimeout:     defaultIdleConnectionTimeout,
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConnsPerHost: defaultMaxIdleConnection,
	}
	// defaultRetryStatusCodes lists http status code for which retry is attempted
	defaultRetryStatusCodes = []int{
		http.StatusTooManyRequests,
		http.StatusRequestTimeout,
		http.StatusInsufficientStorage,
		http.StatusGatewayTimeout,
		http.StatusServiceUnavailable,
	}
)

// CheckRetryRequired - specifies generic condition required for retry
// it takes http response status code as input and return boolean value
// if true http request module will retry
type CheckRetryRequired func(int) bool

// standard retry check function
var (
	// ExactResponseCodeMatch will only retry if response code matches the default retry codes
	// This is the default method if no retryCondition is specified by the application
	ExactResponseCodeMatch CheckRetryRequired = func(statusCode int) bool {
		retryFlag := false
		for _, retryCode := range defaultRetryStatusCodes {
			retryFlag = retryFlag || (retryCode == statusCode)
		}
		return retryFlag
	}
)

// RequestSpecifications - controls the each http requests behaviour
type RequestSpecifications struct {
	URL            string
	HTTPMethod     string
	RequestType    string
	Headers        map[string]string
	Params         map[string]interface{}
	Timeout        int
	UseAuth        bool
	Username       string
	Password       string
	RetryCount     int
	RetryInterval  int
	RetryCondition CheckRetryRequired
	Log            *logrus.Entry
}

// RequestHandler - holds request handler information
type RequestHandler struct {
	appName string
	Handler *gorequest.SuperAgent
}

// NewRequestHandler  - return RequestHandler object
func NewRequestHandler(app string) *RequestHandler {
	return &RequestHandler{
		appName: app,
		Handler: gorequest.New(),
	}
}

// Fetch - prepare and send HTTP request and return the response.
func (r *RequestHandler) Fetch(specs *RequestSpecifications) (int, []byte, http.Header) {
	statusCode := http.StatusInternalServerError
	var requestCount uint
	var response gorequest.Response // store intermediate go-response object reference
	var body []byte                 // store intermediate response body string
	var err []error                 // store intermediate error slice
	var headers http.Header         // store intermediate headers

	logFields := logrus.Fields{
		"url":     specs.URL,
		"appName": r.appName,
		"module":  "httprequests",
	}
	newHandler := r.prepareRequest(specs, logFields)
	logFields["method"] = specs.HTTPMethod
	requestLog := specs.Log.WithFields(logFields)
	// manually hand the retry operation as it can be useful for logging purpose
	for int(requestCount) <= specs.RetryCount {
		// skip the first loop for retry
		if requestCount >= 1 {
			if !specs.RetryCondition(statusCode) {
				break
			}
			requestLog.Warnf("retry attempt %v out of %v", requestCount, specs.RetryCount)
			interval := time.Duration(specs.RetryInterval) * time.Second
			requestLog.WithFields(logrus.Fields{"retry attempt": specs.RetryCount})
			time.Sleep(interval)
		}
		startTime := GetCurrentTimeString()
		// finally sending the request
		response, body, err = newHandler.EndBytes()
		processResponse(response, err, &statusCode, &headers, specs)
		requestLog.WithFields(logrus.Fields{
			"params":                      specs.Params,
			"header":                      specs.Headers,
			"http_response_code":          statusCode,
			"http_response_body":          string(body),
			"http_time_elapsed (in secs)": fmt.Sprintf("%.4f", GetCurrentDelta(startTime).Seconds()),
		}).Info("http response received")
		// incrementing the request count
		requestCount++
	}
	return statusCode, body, headers
}

// processResponse - update correct statusCode and log errors if any
// timeout: sets correct http response code (408) as in this case response object is nil
// redirect: by default go allows 10 redirect but user can specify it exclusively using
// request specifications or completely disables it.
func processResponse(response gorequest.Response, err []error, statusCode *int, headers *http.Header, specs *RequestSpecifications) {
	requestLog := logrus.New().WithFields(logrus.Fields{
		"url":    specs.URL,
		"method": specs.HTTPMethod,
	})
	//prepare http response code
	if err != nil {
		errMsg := "httprequest:[fetch]"
		//investigates timeout error
		if checkTimeout(err) {
			*statusCode = http.StatusRequestTimeout
			errMsg = fmt.Sprintf("%v timeout encountered with request timeout set to %v seconds", errMsg, specs.Timeout)
		}
		errMsg = fmt.Sprintf("%v with headers %+v and params %+v.", errMsg, specs.Headers, specs.Params)
		errString := GetErrorString(err)
		requestLog.Errorf("%v %v", errMsg, errString)
	} else {
		*statusCode = response.StatusCode
		*headers = response.Header
	}
}

// prepareRequest - return customized request handler below are default values if not exclusively specified
// for RequestSpecifications struct
// HttpMethod          : "GET"
// UseAuth             : false
// RetryInterval       : 1 second if retry count is non-zero
// RetryCondition      : ExactResponseCodeMatch only active if RetryCount > 0
// RequestType         : json
func (r *RequestHandler) prepareRequest(specs *RequestSpecifications, logFields map[string]interface{}) *gorequest.SuperAgent {
	// set transport
	timeout := GetValue(specs.Timeout, defaultTimeout).(int)
	logFields["http_timeout"] = timeout
	r.Handler.Transport = DefaultTransport
	r.Handler.Client.Timeout = time.Duration(timeout) * time.Second

	if specs.RetryCondition == nil {
		specs.RetryCondition = ExactResponseCodeMatch
	}
	// specify authorization for request
	if specs.UseAuth {
		r.Handler = r.Handler.SetBasicAuth(specs.Username, specs.Password)
		logFields["http_require_auth"] = specs.UseAuth
	}
	// checks if request retry is enabled
	if specs.RetryCount > 0 {
		//sets the default retry interval if not specified
		if specs.RetryInterval == 0 {
			specs.RetryInterval = defaultRetryInterval
		}
		logFields["http_retry_count"] = specs.RetryCount
		logFields["http_retry_interval"] = specs.RetryInterval
	}

	// set the default http request method to "GET"
	if len(specs.HTTPMethod) == 0 {
		specs.HTTPMethod = http.MethodGet
	}
	logFields["http_method"] = specs.HTTPMethod
	r.Handler = addHeadersAndBody(specs, r.Handler)
	specs.Log.WithFields(logFields).Info("prepared and sending http request")
	return r.Handler
}

// addHeadersAndBody - adds headers and body to request handler
func addHeadersAndBody(specs *RequestSpecifications, newHandler *gorequest.SuperAgent) *gorequest.SuperAgent {
	requestURL := specs.URL
	if specs.HTTPMethod == http.MethodGet {
		// handle GET request
		newHandler = setRequestType(newHandler.Get(requestURL), specs)
	} else if specs.HTTPMethod == http.MethodPost {
		// handle POST request
		newHandler = setRequestType(newHandler.Post(requestURL), specs)
		// check if post params exists
		if specs.Params != nil && len(specs.Params) > 0 {
			newHandler = newHandler.Send(specs.Params)
		}
	} else if specs.HTTPMethod == http.MethodDelete {
		// handle DELETE request
		newHandler = setRequestType(newHandler.Delete(requestURL), specs)
	}
	// set the header for request
	for headerKey, headerValue := range specs.Headers {
		cleanKey := strings.TrimSpace(headerKey)
		if len(cleanKey) == 0 {
			continue
		}
		cleanValue := strings.TrimSpace(headerValue)
		newHandler.Set(cleanKey, cleanValue)
	}
	return newHandler
}

// CheckTimeout - return true if request times out
func checkTimeout(errSlice []error) bool {
	timeoutFlag := false
	for _, err := range errSlice {
		timeoutFlag = timeoutFlag || os.IsTimeout(err)
	}
	return timeoutFlag
}

// setRequestType - setup the http request body type
func setRequestType(request *gorequest.SuperAgent, specs *RequestSpecifications) *gorequest.SuperAgent {
	requestType := jsonRequestType
	if len(specs.RequestType) > 0 {
		_, exists := gorequest.Types[specs.RequestType]
		// use default request body type as json
		if exists {
			requestType = specs.RequestType
		}
		specs.Log.Debugf("using request body type : %v", requestType)
	}
	return request.Type(requestType)
}
