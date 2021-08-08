package utils

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID(t *testing.T) {
	check := assert.New(t)
	uuid1 := GetTransactionId()
	uuid2 := GetTransactionId()
	check.NotEqual(uuid1, uuid2)
}

func TestGetCurrentTimeString(t *testing.T) {
	check := assert.New(t)
	time1, _ := time.Parse("2006-01-02 15:04:05.000000", GetCurrentTimeString())
	time2 := time.Now().UTC()
	check.WithinDuration(time1, time2, 1*time.Second)
}

func TestPositiveGetCurrentDelta(t *testing.T) {
	check := assert.New(t)
	timeDuration := GetCurrentDelta(GetCurrentTimeString())
	check.NotEqual(timeDuration, time.Duration(ZeroSecond))
}

func TestNegativeGetCurrentDelta(t *testing.T) {
	check := assert.New(t)
	timeDuration := GetCurrentDelta("12345")
	check.Equal(timeDuration, time.Duration(ZeroSecond))
}

func TestGetValue(t *testing.T) {
	check := assert.New(t)
	var tests = []struct {
		input    interface{}
		expected interface{}
	}{
		{GetValue(0, 10), 10},
		{GetValue("", "20"), "20"},
		{GetValue(uint(0), uint(20)), uint(20)},
		{GetValue(nil, nil), nil},
	}
	for _, test := range tests {
		check.Equal(test.input, test.expected)
	}
}

func TestGetTransactionId(t *testing.T) {
	check := assert.New(t)
	id1 := GetTransactionId()
	id2 := GetTransactionId()
	check.NotEqual(id1, id2)
}

func TestGetErrorString(t *testing.T) {
	check := assert.New(t)
	var tests = []struct {
		input    string
		expected string
	}{
		{GetErrorString([]error{errors.New("First error"), errors.New("Second error")}), ". First error. Second error"},
		{GetErrorString(nil), ""},
	}
	for _, test := range tests {
		check.Equal(test.input, test.expected)
	}
}
