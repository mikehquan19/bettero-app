package repositories

import (
	"betterov2/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type BillRepo struct{}

func NewBillRepo() *BillRepo {
	return &BillRepo{}
}

// ListBills gets the list of bills ordered by its due date
func (r *BillRepo) ListBills(ctx context.Context, dbtx DBTX, userId uuid.UUID) ([]models.Bill, error) {
	var bills []models.Bill

	const listBillSQL = `
	SELECT
		b.id,
		CASE WHEN a.id IS NULL THEN NULL ELSE json_build_object(
			'id', a.id,
			'account_number', a.account_number,
			'name', a.name,
			'institution', a.institution,
			'type', a.type
		) END AS account,
		b.merchant,
		b.description,
		b.category,
		b.amount_cents,
		b.due_date
	FROM bills b
	LEFT JOIN accounts a ON b.account_id = a.id
	WHERE b.user_id = $1
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
func (r *BillRepo) GetBill(ctx context.Context, dbtx DBTX, userId, id uuid.UUID) (models.Bill, error) {
	var bill models.Bill

	const getBillSQL = `
	SELECT
		b.id,
		CASE WHEN a.id IS NULL THEN NULL ELSE json_build_object(
			'id', a.id,
			'account_number', a.account_number,
			'name', a.name,
			'institution', a.institution,
			'type', a.type
		) END AS account,
		b.merchant,
		b.description,
		b.category,
		b.amount_cents,
		b.due_date
	FROM bills b
	LEFT JOIN accounts a ON b.account_id = a.id
	WHERE b.user_id = $1 AND b.id = $2;`

	row := dbtx.QueryRow(ctx, getBillSQL, userId, id)
	if err := scanBill(row, &bill); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Bill{}, models.ErrNotFound
		}
		return models.Bill{}, err
	}

	return bill, nil
}

// InsertBill inserts a bill and returns the bill of the account
func (r *BillRepo) InsertBill(ctx context.Context, dbtx DBTX, userId uuid.UUID, body models.BillBody) (models.Bill, error) {
	var newBill models.Bill

	const insertBillSQL = `
	WITH new_bill AS (
		INSERT INTO bills (
			user_id,
			account_id,
			merchant,
			description,
			category,
			amount_cents,
			due_date
		)
		SELECT $1, $2, $3, $4, $5, $6, $7
		WHERE EXISTS (SELECT 1 FROM accounts a WHERE a.id = $2 AND a.user_id = $1)
		RETURNING *
	)
	SELECT
		b.id,
		CASE WHEN a.id IS NULL THEN NULL ELSE json_build_object(
			'id', a.id,
			'account_number', a.account_number,
			'name', a.name,
			'institution', a.institution,
			'type', a.type
		) END AS account,
		b.merchant,
		b.description,
		b.category,
		b.amount_cents,
		b.due_date
	FROM new_bill b
	LEFT JOIN accounts a ON b.account_id = a.id;`

	row := dbtx.QueryRow(ctx, insertBillSQL,
		userId,
		body.AccountID,
		body.Merchant,
		body.Description,
		body.Category,
		body.AmountCents,
		body.DueDate,
	)
	if err := scanBill(row, &newBill); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Bill{}, models.ErrForeignKey
		}
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
	userId uuid.UUID,
	id uuid.UUID,
	body models.BillBody,
) (models.Bill, error) {
	var updatedBill models.Bill

	const updateBillSQL = `
	WITH updated_bill AS (
		UPDATE bills b
		SET account_id = $3,
			merchant = $4,
			description = $5,
			category = $6,
			amount_cents = $7,
			due_date = $8
		WHERE b.id = $1 AND b.user_id = $2 AND EXISTS (SELECT 1 FROM accounts a WHERE a.id = $3 AND a.user_id = $2)
		RETURNING *
	)
	SELECT
		b.id,
		CASE WHEN a.id IS NULL THEN NULL ELSE json_build_object(
			'id', a.id,
			'account_number', a.account_number,
			'name', a.name,
			'institution', a.institution,
			'type', a.type
		) END AS account,
		b.merchant,
		b.description,
		b.category,
		b.amount_cents,
		b.due_date
	FROM updated_bill b
	LEFT JOIN accounts a ON b.account_id = a.id;`

	row := dbtx.QueryRow(ctx, updateBillSQL,
		id,
		userId,
		body.AccountID,
		body.Merchant,
		body.Description,
		body.Category,
		body.AmountCents,
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
func (r *BillRepo) DeleteBill(ctx context.Context, dbtx DBTX, userId, id uuid.UUID) (models.Bill, error) {
	var deletedBill models.Bill

	const deleteBillSQL = `
	WITH deleted_bill AS (
		DELETE FROM bills WHERE id = $1 AND user_id = $2 RETURNING *
	)
	SELECT
		b.id,
		CASE WHEN a.id IS NULL THEN NULL ELSE json_build_object(
			'id', a.id,
			'account_number', a.account_number,
			'name', a.name,
			'institution', a.institution,
			'type', a.type
		) END AS account,
		b.merchant,
		b.description,
		b.category,
		b.amount_cents,
		b.due_date
	FROM deleted_bill b
	LEFT JOIN accounts a ON b.account_id = a.id;`

	row := dbtx.QueryRow(ctx, deleteBillSQL, id, userId)
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
		&bill.AmountCents,
		&bill.DueDate,
	)
}
