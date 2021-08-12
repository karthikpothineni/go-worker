package dataadapters

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"go-worker/logger"
	"go-worker/models"
)

// TestPositiveUpdation - test a successful db update
func TestPositiveUpdation(t *testing.T) {

	// Init logger
	logger.Init()

	// create sqlmock object
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mysqlDB, err = gorm.Open("mysql", db)
	defer db.Close()

	// mock update query
	mock.ExpectExec("UPDATE call_info SET billing_cost = ?").
		WithArgs("1.2", "e21b0dda-6566-402a-8f8c-0657e5b87eeb").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// call data adapter update function
	balanceResponse := getBalanceResponse()
	if err = UpdateCallInfo(balanceResponse); err != nil {
		t.Errorf("Error was not expected while updating cost: %s", err)
	}

	// check mock result
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

// TestNegativeUpdation - test a failure db update
func TestNegativeUpdation(t *testing.T) {

	check := assert.New(t)
	// Init logger
	logger.Init()

	// create sqlmock object
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	mysqlDB, err = gorm.Open("mysql", db)
	defer db.Close()

	// mock update query
	mock.ExpectExec("UPDATE call_info SET billing_cost = ?").
		WithArgs("1.1", "d12b0dda-6566-402a-8f8c-0657e5b87eeb").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// call data adapter update function
	balanceResponse := getBalanceResponse()
	err = UpdateCallInfo(balanceResponse)
	check.Contains(err.Error(), "ExecQuery")
}

// getBalanceResponse - returns balance response information
func getBalanceResponse() (balanceResponse models.BalanceResponse) {
	balanceResponse = models.BalanceResponse{
		ChargeAmount:       "1.2",
		CallID:             "e21b0dda-6566-402a-8f8c-0657e5b87eeb",
		TotalConsumedUnits: 2,
	}
	return
}
