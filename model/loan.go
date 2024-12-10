package model

import (
	"database/sql"
	"time"
)

type Investment struct {
	InvestorID string    `json:"investor_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Amount     float64   `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
}

type Approval struct {
	FieldValidatorID sql.NullString `json:"field_validator_id"`
	ProofImageURL    sql.NullString `json:"proof_image_url"`
	ApprovalDate     sql.NullTime   `json:"approval_date"`
}

type Disbursement struct {
	FieldOfficerID           sql.NullString `json:"field_officer_id"`
	SignedAgreementLetterURL sql.NullString `json:"signed_agreement_letter_url"`
	DisbursementDate         sql.NullTime   `json:"disbursement_date"`
}

type Transition struct {
	LoanID        string    `json:"loan_id"`
	PreviousState LoanState `json:"previous_state"`
	Event         LoanEvent `json:"event"`
	NextState     LoanState `json:"next_state"`
}

type Loan struct {
	ID                    string    `json:"id"`
	BorrowerID            string    `json:"borrower_id"`
	PrincipalAmount       float64   `json:"principal_amount"`
	Rate                  float64   `json:"rate"`
	ROI                   float64   `json:"roi"`
	State                 LoanState `json:"state"`
	TotalInvestmentAmount float64   `json:"total_investment_amount,omitempty"`
	Version               int       `json:"version"`

	AgreementLetterURL sql.NullString `json:"agreement_letter_url"`
	NewInvestment      Investment     `json:"new_investment,omitempty"`
	Approval           Approval       `json:"approval,omitempty"`
	Investments        []Investment   `json:"investments,omitempty"`
	Disbursement       Disbursement   `json:"disbursement,omitempty"`
}

func (a *Loan) SetAgreementURL(agreementURL string) {
	a.AgreementLetterURL = sql.NullString{String: agreementURL, Valid: true}
}

type CreateLoanRequest struct {
	BorrowerID      string  `json:"borrower_id"`
	PrincipalAmount float64 `json:"principal_amount"`
	Rate            float64 `json:"rate"`
	ROI             float64 `json:"roi"`
}

type ApproveLoanRequest struct {
	ValidatorID   string    `json:"validator_id"`
	ProofImageURL string    `json:"proof_image_url"`
	ApprovalDate  time.Time `json:"approval_date"`
	LoanID        string
}

func (a *ApproveLoanRequest) ToApproval() Approval {
	return Approval{
		FieldValidatorID: sql.NullString{String: a.ValidatorID, Valid: true},
		ProofImageURL:    sql.NullString{String: a.ProofImageURL, Valid: true},
		ApprovalDate:     sql.NullTime{Time: a.ApprovalDate, Valid: true},
	}
}

type AddInvestmentRequest struct {
	InvestorID string  `json:"investor_id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Amount     float64 `json:"amount"`
	LoanID     string
}

func (a *AddInvestmentRequest) ToInvestment() Investment {
	return Investment{
		InvestorID: a.InvestorID,
		Email:      a.Email,
		Name:       a.Name,
		Amount:     a.Amount,
	}
}

type DisburseLoanRequest struct {
	OfficerID          string    `json:"officer_id"`
	AgreementLetterURL string    `json:"agreement_letter_url"`
	DisbursementDate   time.Time `json:"disbursement_date"`
	LoanID             string
}

func (a *DisburseLoanRequest) ToDisbursement() Disbursement {
	return Disbursement{
		FieldOfficerID:           sql.NullString{String: a.OfficerID, Valid: true},
		SignedAgreementLetterURL: sql.NullString{String: a.AgreementLetterURL, Valid: true},
		DisbursementDate:         sql.NullTime{Time: a.DisbursementDate, Valid: true},
	}
}

func isStringNullOrEmpty(ns sql.NullString) bool {
	if !ns.Valid || ns.String == "" {
		return true // NULL or empty
	}
	return false // Has a non-empty value
}

func isTimeNullOrEmpty(nt sql.NullTime) bool {
	return !nt.Valid || nt.Time.IsZero()
}
