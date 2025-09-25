package main

import (
	"database/sql"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type dbPostgres struct {
	db *sql.DB
}

func (db *dbPostgres) createPayment(pr *cmn.PaymentRequest) error {
	// TODO: check affected row count == 1
	_, err := db.db.Exec(`
	INSERT INTO payments.transfer (
		system_id,
		app_id,
		source_account_id,
		target_account_id,
		amount,
		status
		)
		VALUES ($1, $2, $3, $4, $5, 'PENDING')
		`, pr.SystemID, pr.AppID, pr.SourceAccountID, pr.TargetAccountID, pr.Amount)

	return err
}
