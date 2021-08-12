package utils

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// MicroTimeStamp - format generates time with YYYY-MM-DD HH:MM:SS.%f
	MicroTimeStamp = "2006-01-02 15:04:05.000000"
	// ZeroSecond - for depicting zero second
	ZeroSecond = 0
)

// GetCurrentTimeString - returns string representation of time in micro timestamp format
func GetCurrentTimeString() string {
	currentTime := time.Now().UTC()
	return currentTime.Format(MicroTimeStamp)
}

// GetTimeDelta - returns time duration between start time and present time
func GetCurrentDelta(startTime string) time.Duration {
	currentTime := time.Now().UTC()
	publishTime, err := time.Parse(MicroTimeStamp, startTime)
	if err != nil {
		return time.Duration(ZeroSecond)
	}
	return currentTime.Sub(publishTime)
}

// GetValue - takes the first parameter as the actual value and second parameter as default value
// it checks if actual value is empty then assigns the default value
func GetValue(value interface{}, defaultValue interface{}) interface{} {
	if reflect.ValueOf(value).Kind() != reflect.ValueOf(defaultValue).Kind() {
		return value
	}
	switch t := value.(type) {
	case string:
		if len(strings.TrimSpace(t)) == 0 {
			return defaultValue
		}
	case uint:
		if t == 0 {
			return defaultValue
		}
	case int:
		if t == 0 {
			return defaultValue
		}
	}
	return value
}

//GetTransactionID - generates a new UUID for a transaction
func GetTransactionID() string {
	return uuid.New().String()
}

//GetErrorString - creates a single error string
func GetErrorString(errs []error) string {
	errString := ""
	for _, err := range errs {
		errString = fmt.Sprintf("%v. %v", errString, err.Error())
	}
	return errString
}
