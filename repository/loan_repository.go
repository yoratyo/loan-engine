package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"loan-engine/model"
	"log"
	"time"
)

type LoanRepositoryInterface interface {
	Create(ctx context.Context, loan *model.Loan) (string, error)
	CreateInvestment(ctx context.Context, loan *model.Loan) error
	CreateTransition(ctx context.Context, t *model.Transition) error
	Update(ctx context.Context, loan *model.Loan) error
	GetLoan(ctx context.Context, id string) (*model.Loan, error)
	GetInvestments(ctx context.Context, loanID string) ([]model.Investment, error)
	WithTransaction(ctx context.Context, fn func(rTx LoanRepositoryInterface) error) error
}

type LoanRepository struct {
	db *sql.DB
	tx *sql.Tx // Active transaction, if any
}

// dbExecutor defines an interface implemented by both *sql.DB and *sql.Tx.
type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func NewLoanRepository(db *sql.DB) LoanRepositoryInterface {
	return &LoanRepository{db: db}
}

// getDB returns the active transaction if any, otherwise the regular DB.
func (r *LoanRepository) getDB() dbExecutor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *LoanRepository) Create(ctx context.Context, loan *model.Loan) (string, error) {
	query := `
        INSERT INTO loans (
            id, borrower_id, principal_amount, rate, roi, state
        ) VALUES (gen_random_uuid(), $1, $2, $3, $4, $5) RETURNING id
    `

	var newID string
	row := r.getDB().QueryRowContext(ctx, query,
		loan.BorrowerID, loan.PrincipalAmount,
		loan.Rate, loan.ROI, loan.State,
	)

	// Scan the row to retrieve the ID
	err := row.Scan(&newID)
	if err != nil {
		return "", err
	}

	return newID, nil
}

func (r *LoanRepository) CreateInvestment(ctx context.Context, loan *model.Loan) error {
	query := `
        INSERT INTO loan_investments (
            id, loan_id, investor_id, investor_name, email, amount
        ) VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
    `
	_, err := r.getDB().ExecContext(ctx, query,
		loan.ID, loan.NewInvestment.InvestorID, loan.NewInvestment.Name, loan.NewInvestment.Email, loan.NewInvestment.Amount,
	)

	return err
}

func (r *LoanRepository) CreateTransition(ctx context.Context, t *model.Transition) error {
	query := `
        INSERT INTO loan_state_transitions (
            loan_id, previous_state, event, next_state
        ) VALUES ($1, $2, $3, $4)
    `
	_, err := r.getDB().ExecContext(ctx, query,
		t.LoanID, t.PreviousState, t.Event, t.NextState,
	)

	return err
}

func (r *LoanRepository) Update(ctx context.Context, loan *model.Loan) error {
	query := `
        UPDATE loans SET
            total_investment_amount = total_investment_amount + $1, 
			state = $2, 
			field_validator_id = $3,
			proof_image_url = $4,
			approval_date = $5,
			agreement_letter_url = $6,
			field_officer_id = $7,
			signed_agreement_letter_url = $8,
			disbursement_date = $9,
			version = version + 1
        WHERE id = $10 AND version = $11
    `

	res, err := r.getDB().ExecContext(ctx, query,
		loan.NewInvestment.Amount, loan.State,
		loan.Approval.FieldValidatorID, loan.Approval.ProofImageURL, loan.Approval.ApprovalDate, loan.AgreementLetterURL,
		loan.Disbursement.FieldOfficerID, loan.Disbursement.SignedAgreementLetterURL, loan.Disbursement.DisbursementDate,
		loan.ID, loan.Version,
	)
	if err != nil {
		return err
	}

	// Optimistic Locking with versioning
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 1 {
		// Successfully updated
		return nil
	}

	return errors.New("loan not updated")
}

func (r *LoanRepository) GetLoan(ctx context.Context, id string) (*model.Loan, error) {
	query := `
        SELECT 
            id, borrower_id, principal_amount, total_investment_amount, rate, roi, state,
			field_validator_id, proof_image_url, approval_date, agreement_letter_url, field_officer_id,
            signed_agreement_letter_url, disbursement_date, version
        FROM loans WHERE id = $1
    `

	loan := &model.Loan{}

	err := r.getDB().QueryRowContext(ctx, query, id).Scan(
		&loan.ID, &loan.BorrowerID, &loan.PrincipalAmount, &loan.TotalInvestmentAmount,
		&loan.Rate, &loan.ROI, &loan.State, &loan.Approval.FieldValidatorID, &loan.Approval.ProofImageURL,
		&loan.Approval.ApprovalDate, &loan.AgreementLetterURL, &loan.Disbursement.FieldOfficerID, &loan.Disbursement.SignedAgreementLetterURL,
		&loan.Disbursement.DisbursementDate, &loan.Version,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("loan not found")
		}
		return nil, err
	}

	return loan, nil
}

func (r *LoanRepository) GetInvestments(ctx context.Context, loanID string) ([]model.Investment, error) {
	query := `
        SELECT investor_id, investor_name, email, amount 
        FROM loan_investments 
        WHERE loan_id = $1
        ORDER BY created_at
    `

	rows, err := r.getDB().QueryContext(ctx, query, loanID)
	if err != nil {
		return nil, fmt.Errorf("error querying investments: %w", err)
	}
	defer rows.Close()

	var investments []model.Investment
	for rows.Next() {
		var inv model.Investment
		err := rows.Scan(
			&inv.InvestorID,
			&inv.Name,
			&inv.Email,
			&inv.Amount,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning investment row: %w", err)
		}
		investments = append(investments, inv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating investment rows: %w", err)
	}

	return investments, nil
}

func (r *LoanRepository) WithTransaction(ctx context.Context, fn func(rTx LoanRepositoryInterface) error) error {
	startTime := time.Now()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		duration := time.Since(startTime)
		log.Printf("Transaction completed in %v", duration)

		if p := recover(); p != nil {
			tx.Rollback()
			log.Println("Transaction rolled back due to panic")
			panic(p)
		} else if err != nil {
			tx.Rollback()
			log.Println("Transaction rolled back due to error:", err)
		} else {
			err = tx.Commit()
			log.Println("Transaction committed successfully")
		}

	}()

	err = fn(&LoanRepository{db: r.db, tx: tx})
	return err
}
