package workerpool

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/sirupsen/logrus"

	"go-worker/config"
	dataAdapters "go-worker/data_adapters"
	"go-worker/externals"
	"go-worker/logger"
	"go-worker/models"
	"go-worker/utils"
)

// Worker - holds worker related information
type Worker struct {
	workerID              int
	SQSClient             sqsiface.SQSAPI
	SQSURL                string
	SQSRetry              int
	MaxEvents             int64
	WaitTime              int64
	BalanceRequestHandler *externals.BalanceRequestHandler
	Log                   *logrus.Entry
}

// NewWorker - returns a new object for Worker
func NewWorker(workerID int) *Worker {
	cfg := config.GetConfig()

	region := cfg.GetString("sqs.region")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		logger.Log.Error("Unable to create aws session")
	}
	// new sqs client
	svc := sqs.New(sess)

	prefix := fmt.Sprintf("WorkerID:%d", workerID)
	transaction := fmt.Sprintf("%v", utils.GetTransactionID())
	log := logger.Log.WithFields(logrus.Fields{"prefix": prefix, "transaction": transaction})
	balanceRequesthandler := externals.NewBalanceRequestHandler(log)

	worker := &Worker{
		workerID:              workerID,
		SQSClient:             svc,
		SQSURL:                cfg.GetString("sqs.url"),
		SQSRetry:              cfg.GetInt("sqs.retry_count"),
		MaxEvents:             cfg.GetInt64("worker.max_events"),
		WaitTime:              cfg.GetInt64("worker.wait_time"),
		BalanceRequestHandler: balanceRequesthandler,
		Log:                   log,
	}
	return worker
}

// Init - starts the worker
func (worker *Worker) Init() {
	go func() {
		for {
			select {
			case <-closeChan:
				worker.Close()
				wg.Done()
				return
			default:
				//run job
				worker.run()
			}
		}
	}()
}

// run - fetches the job from SQS and process it
func (worker *Worker) run() {

	// catch panics
	defer func() {
		if err := recover(); err != nil {
			worker.Log.WithField("error", err).Error("Recovering worker after panic")
		}
	}()

	// fetch the job from SQS
	sqsResponse, err := worker.fetch()
	if err != nil {
		logger.Log.WithError(err).Error("Unable to fetch messages from SQS")
		return
	}

	// process sqs messages
	worker.processSQSMessages(sqsResponse)
}

// fetch - is used for de queuing the job from SharQ
func (worker *Worker) fetch() (*sqs.ReceiveMessageOutput, error) {
	result, err := worker.SQSClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},

		QueueUrl:            &worker.SQSURL,
		MaxNumberOfMessages: aws.Int64(worker.MaxEvents),
		WaitTimeSeconds:     aws.Int64(worker.WaitTime),
	})
	return result, err
}

// processSQSMessages - processes sqs messages sequentially
func (worker *Worker) processSQSMessages(sqsResponse *sqs.ReceiveMessageOutput) {

	if len(sqsResponse.Messages) == 0 {
		logger.Log.Info("SQS response is nil")
		return
	}
	var requestIDList []*sqs.DeleteMessageBatchRequestEntry
	var isSuccess bool

	// process each billing event
	for _, message := range sqsResponse.Messages {
		billingEvent := models.BillingEvent{}
		bytesStr := []byte(*message.Body)
		err := json.Unmarshal(bytesStr, &billingEvent)
		if err != nil {
			logger.Log.WithError(err).WithField("bytesStr", bytesStr).Info("Error while unmarshalling sqs message")
			continue
		}
		isSuccess = worker.processBillingEvent(billingEvent)
		if isSuccess {
			deleteMessageRequestEntry := sqs.DeleteMessageBatchRequestEntry{}
			deleteMessageRequestEntry.ReceiptHandle = message.ReceiptHandle
			deleteMessageRequestEntry.Id = message.MessageId
			requestIDList = append(requestIDList, &deleteMessageRequestEntry)
		}
	}

	// delete sqs messages
	worker.deleteSQSMessages(requestIDList)
}

// processBillingEvent - processes bill event
func (worker *Worker) processBillingEvent(billEvent models.BillingEvent) bool {

	// call balance api
	response, isSuccessful := worker.BalanceRequestHandler.BillUser(billEvent)
	if isSuccessful {
		var balanceResponse models.BalanceResponse
		err := json.Unmarshal(response, &balanceResponse)
		if err != nil {
			logger.Log.WithError(err).WithField("call_id: ", billEvent.CallID).Info("Unable to map json response to struct")
			return false
		}
		// update call info in database
		updateErr := dataAdapters.UpdateCallInfo(balanceResponse)
		if updateErr != nil {
			return false
		}
		return true
	}
	return false
}

// deleteSQSMessages - deletes sqs messages once processed
func (worker *Worker) deleteSQSMessages(requestIDList []*sqs.DeleteMessageBatchRequestEntry) {
	if len(requestIDList) > 0 {
		for i := 0; i < worker.SQSRetry; i++ {
			resp, err := worker.SQSClient.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
				QueueUrl: &worker.SQSURL,
				Entries:  requestIDList,
			})
			if err != nil {
				logger.Log.WithError(err).Info("Unable to delete messages from sqs")
				if len(resp.Failed) > 0 {
					for _, failedDelete := range resp.Failed {
						logger.Log.WithFields(logrus.Fields{
							"Code":        failedDelete.Code,
							"Id":          failedDelete.Id,
							"Message":     failedDelete.Message,
							"SenderFault": failedDelete.SenderFault,
						}).Info("Error while deleting sqs message")
					}
				}
				continue
			}
			break
		}
	}
}

//Close uninitializes a worker
func (worker *Worker) Close() {
	worker.Log.Infof("Successfully closed worker %d", worker.workerID)
}
