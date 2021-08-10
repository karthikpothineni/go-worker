package workerpool

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var queueURL = "https://queue.amazonaws.com/88888EXAMPLE/MyQueue"

// SQSMessage - sqs test message
const SQSMessage = "Test SQS!"

// mockSQS - holds sqs mocking info
type mockSQS struct {
	sqsiface.SQSAPI
	messages map[string][]*sqs.Message
}

// TestPositiveFetch - tests a successful sqs fetch
func TestPositiveFetch(t *testing.T) {
	q, worker := getMockClient()
	_, _ = q.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(SQSMessage),
		QueueUrl:    &queueURL,
	})

	// call worker fetch
	message, _ := worker.fetch()
	assert.Equal(t, *message.Messages[0].Body, SQSMessage)
}

// TestNegativeFetch - tests a failure sqs fetch
func TestNegativeFetch(t *testing.T) {
	q, worker := getMockClient()
	_, _ = q.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String("Test SQSiface!"),
		QueueUrl:    &queueURL,
	})

	// call worker fetch
	message, _ := worker.fetch()
	assert.NotEqual(t, *message.Messages[0].Body, SQSMessage)
}

// TestPositiveDeleteSQSMessages - tests a successful sqs delete
func TestPositiveDeleteSQSMessages(t *testing.T) {
	var requestIDList []*sqs.DeleteMessageBatchRequestEntry
	q, worker := getMockClient()

	// send sqs message
	_, _ = q.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(SQSMessage),
		QueueUrl:    &queueURL,
	})

	// delete sqs message
	deleteMessageRequestEntry := sqs.DeleteMessageBatchRequestEntry{}
	deleteMessageRequestEntry.ReceiptHandle = nil
	deleteMessageRequestEntry.Id = nil
	requestIDList = append(requestIDList, &deleteMessageRequestEntry)
	worker.deleteSQSMessages(requestIDList)

	// fetch sqs message
	message, _ := worker.fetch()
	assert.Equal(t, len(message.Messages), 0)
}

// TestNegativeDeleteSQSMessages - tests a failure sqs delete
func TestNegativeDeleteSQSMessages(t *testing.T) {
	q, worker := getMockClient()

	// send sqs message
	_, _ = q.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(SQSMessage),
		QueueUrl:    &queueURL,
	})

	// delete sqs message
	worker.deleteSQSMessages(nil)

	// fetch sqs message
	message, _ := worker.fetch()
	assert.NotEqual(t, len(message.Messages), 0)
}

// getMockClient - returns mocked sqs, worker
func getMockClient() (sqsiface.SQSAPI, *Worker) {
	mocksqs := &mockSQS{
		messages: map[string][]*sqs.Message{},
	}
	log := logrus.New().WithFields(logrus.Fields{"test_worker": 1})
	worker := &Worker{
		SQSClient: mocksqs,
		SQSURL:    queueURL,
		SQSRetry:  1,
		Log:       log,
	}

	return mocksqs, worker
}

// SendMessage - mock function for sending message
func (m *mockSQS) SendMessage(in *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	m.messages[*in.QueueUrl] = append(m.messages[*in.QueueUrl], &sqs.Message{
		Body: in.MessageBody,
	})
	return &sqs.SendMessageOutput{}, nil
}

// ReceiveMessage - mock function for receiving messages
func (m *mockSQS) ReceiveMessage(in *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	if len(m.messages[*in.QueueUrl]) == 0 {
		return &sqs.ReceiveMessageOutput{}, nil
	}
	response := m.messages[*in.QueueUrl][0:1]
	m.messages[*in.QueueUrl] = m.messages[*in.QueueUrl][1:]
	return &sqs.ReceiveMessageOutput{
		Messages: response,
	}, nil
}

// DeleteMessageBatch - mock function for deleting messages
func (m *mockSQS) DeleteMessageBatch(de *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error) {
	if len(m.messages[*de.QueueUrl]) == 0 {
		return &sqs.DeleteMessageBatchOutput{}, nil
	}
	m.messages[*de.QueueUrl] = m.messages[*de.QueueUrl][1:]
	return &sqs.DeleteMessageBatchOutput{
		Failed: nil,
	}, nil
}
