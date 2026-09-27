package repositories

import (
	"betterov2/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type BillRepo struct{}

func NewBillRepo() *BillRepo {
	return &BillRepo{}
}

// ListBills gets the list of bills ordered by its due date
func (r *BillRepo) ListBills(ctx context.Context, dbtx DBTX, userId int64) ([]models.Bill, error) {
	var bills []models.Bill

	const listBillSQL = `
	SELECT
		b.id, 
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		b.merchant, 
		b.description, 
		b.category, 
		b.amount, 
		b.due_date
	FROM bills b 
	JOIN accounts a ON b.account_id = a.id
	WHERE a.user_id = $1
	ORDER BY b.due_date ASC;`

	rows, err := dbtx.Query(ctx, listBillSQL, userId)
	if err != nil {
		return nil, err
	}

	bills, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Bill])
	if err != nil {
		return nil, err
	}

	return bills, nil
}

// GetBill returns the list of bills
func (r *BillRepo) GetBill(ctx context.Context, dbtx DBTX, id int64) (models.Bill, error) {
	var bill models.Bill

	const getBillSQL = `
	SELECT
		b.id,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		b.merchant, 
		b.description, 
		b.category, 
		b.amount, 
		b.due_date
	FROM bills b
	JOIN accounts a ON b.account_id = a.id
	WHERE b.id = $1;`

	row := dbtx.QueryRow(ctx, getBillSQL, id)
	if err := scanBill(row, &bill); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Bill{}, models.ErrNotFound
		}
		return models.Bill{}, err
	}

	return bill, nil
}

// InsertBill inserts a bill and returns the bill of the account
func (r *BillRepo) InsertBill(ctx context.Context, dbtx DBTX, body models.BillBody) (models.Bill, error) {
	var newBill models.Bill

	const insertBillSQL = `
	WITH new_bill AS (
		INSERT INTO bills (
			account_id, 
			merchant, 
			description, 
			category, 
			amount, 
			due_date
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *
	)
	SELECT
		b.id,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		b.merchant, 
		b.description, 
		b.category, 
		b.amount, 
		b.due_date
	FROM new_bill b
	JOIN accounts a ON b.account_id = a.id;`

	row := dbtx.QueryRow(ctx, insertBillSQL,
		body.AccountID,
		body.Merchant,
		body.Description,
		body.Category,
		body.Amount,
		body.DueDate,
	)
	if err := scanBill(row, &newBill); err != nil {
		if isForeignKeyViolation(err) {
			return models.Bill{}, models.ErrForeignKey
		}
		return models.Bill{}, err
	}

	return newBill, nil
}

// UpdateBill updates and returns the bill
func (r *BillRepo) UpdateBill(
	ctx context.Context,
	dbtx DBTX,
	id int64,
	body models.BillBody,
) (models.Bill, error) {
	var updatedBill models.Bill

	const updateBillSQL = `
	WITH updated_bill AS (
		UPDATE bills
		SET account_id = $2, 
			merchant = $3, 
			description = $4, 
			category = $5, 
			amount = $6, 
			due_date = $7
		WHERE id = $1
		RETURNING *
	)
	SELECT
		b.id,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		b.merchant, 
		b.description, 
		b.category,
		b.amount, 
		b.due_date
	FROM updated_bill b
	JOIN accounts a ON b.account_id = a.id;`

	row := dbtx.QueryRow(ctx, updateBillSQL,
		id,
		body.AccountID,
		body.Merchant,
		body.Description,
		body.Category,
		body.Amount,
		body.DueDate,
	)
	if err := scanBill(row, &updatedBill); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Bill{}, models.ErrNotFound
		}
		return models.Bill{}, err
	}

	return updatedBill, nil
}

// DeleteBill deletes and returns the bill by ID
func (r *BillRepo) DeleteBill(ctx context.Context, dbtx DBTX, id int64) (models.Bill, error) {
	var deletedBill models.Bill

	const deleteBillSQL = `
	WITH deleted_bill AS (
		DELETE FROM bills WHERE id = $1 RETURNING *
	)
	SELECT
		b.id,
		json_build_object(
			'id', a.id,
			'acc_number', a.acc_number,
			'acc_name', a.acc_name,
			'institution', a.institution,
			'type', a.type
		) AS account,
		b.merchant, 
		b.description, 
		b.category, 
		b.amount, 
		b.due_date
	FROM deleted_bill b
	JOIN accounts a ON b.account_id = a.id;`

	row := dbtx.QueryRow(ctx, deleteBillSQL, id)
	if err := scanBill(row, &deletedBill); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Bill{}, models.ErrNotFound
		}
		return models.Bill{}, err
	}

	return deletedBill, nil
}

// scanBill parses the returned row into bill
func scanBill(billRow pgx.Row, bill *models.Bill) error {
	return billRow.Scan(
		&bill.ID,
		&bill.Account,
		&bill.Merchant,
		&bill.Description,
		&bill.Category,
		&bill.Amount,
		&bill.DueDate,
	)
}
