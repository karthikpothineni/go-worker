package dataAdapters

import (
	"go-worker/logger"
	"go-worker/models"
)

// UpdateCallInfo - updates the billing cost for a call
func UpdateCallInfo(balanceResponse models.BalanceResponse) error {

	query := "UPDATE call_info SET billing_cost = ? WHERE call_id = ?;"
	err := mysqlDB.Exec(query, balanceResponse.ChargeAmount, balanceResponse.CallID).Error
	if err != nil {
		logger.Log.WithError(err).WithField("call_id: ", balanceResponse.CallID).Info("Unable to update billing info")
	}

	return err
}
