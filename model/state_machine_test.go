package model_test

import (
	"database/sql"
	"loan-engine/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createValidLoan() *model.Loan {
	return &model.Loan{
		ID:              "loan-123",
		BorrowerID:      "borrower-123",
		PrincipalAmount: 1000.0,
		Rate:            5.0,
		ROI:             10.0,
		State:           model.StateInitial,
		Version:         1,
	}
}

func TestStateMachineTransitions(t *testing.T) {
	tests := []struct {
		name          string
		loan          *model.Loan
		event         model.LoanEvent
		setupFn       func(*model.Loan)
		expectedState model.LoanState
		expectError   bool
		errorMessage  string
	}{
		{
			name:  "Valid Initial to Proposed Transition",
			loan:  createValidLoan(),
			event: model.EventSubmission,
			setupFn: func(l *model.Loan) {
				// Valid loan already set up
			},
			expectedState: model.StateProposed,
			expectError:   false,
		},
		{
			name:  "Invalid Initial to Proposed - Empty BorrowerID",
			loan:  createValidLoan(),
			event: model.EventSubmission,
			setupFn: func(l *model.Loan) {
				l.BorrowerID = ""
			},
			expectedState: model.StateInitial,
			expectError:   true,
			errorMessage:  "state decision failed: loan borrower ID data is empty",
		},
		{
			name:  "Valid Proposed to Approved Transition",
			loan:  createValidLoan(),
			event: model.EventApprove,
			setupFn: func(l *model.Loan) {
				l.State = model.StateProposed
				l.Approval = model.Approval{
					FieldValidatorID: sql.NullString{String: "validator-123", Valid: true},
					ProofImageURL:    sql.NullString{String: "http://example.com/proof.jpg", Valid: true},
					ApprovalDate:     sql.NullTime{Time: time.Now(), Valid: true},
				}
			},
			expectedState: model.StateApproved,
			expectError:   false,
		},
		{
			name:  "Valid Approved to Invested Transition",
			loan:  createValidLoan(),
			event: model.EventAddInvestment,
			setupFn: func(l *model.Loan) {
				l.State = model.StateApproved
				l.NewInvestment = model.Investment{
					InvestorID: "investor-123",
					Name:       "John Doe",
					Email:      "john@example.com",
					Amount:     1000.0,
				}
			},
			expectedState: model.StateInvested,
			expectError:   false,
		},
		{
			name:  "Valid Invested to Disbursed Transition",
			loan:  createValidLoan(),
			event: model.EventDisburseFunds,
			setupFn: func(l *model.Loan) {
				l.State = model.StateInvested
				l.Disbursement = model.Disbursement{
					FieldOfficerID:           sql.NullString{String: "officer-123", Valid: true},
					SignedAgreementLetterURL: sql.NullString{String: "http://example.com/agreement.pdf", Valid: true},
					DisbursementDate:         sql.NullTime{Time: time.Now(), Valid: true},
				}
			},
			expectedState: model.StateDisbursed,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the loan for the test case
			tt.setupFn(tt.loan)

			// Create new state machine for each test
			sm := model.NewStateMachine(tt.loan.State)

			// Attempt transition
			err := sm.Transition(tt.loan, tt.event)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Equal(t, tt.errorMessage, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedState, tt.loan.State)
			assert.Equal(t, tt.expectedState, sm.GetCurrentState())
		})
	}
}

func TestAddInvestmentRulePartialInvestment(t *testing.T) {
	loan := createValidLoan()
	loan.State = model.StateApproved
	loan.NewInvestment = model.Investment{
		InvestorID: "investor-123",
		Name:       "John Doe",
		Email:      "john@example.com",
		Amount:     500.0, // Half of principal amount
	}

	sm := model.NewStateMachine(loan.State)
	err := sm.Transition(loan, model.EventAddInvestment)

	assert.NoError(t, err)
	assert.Equal(t, model.StateApproved, loan.State) // Should stay in Approved state
}

func TestAddInvestmentRuleExceedPrincipal(t *testing.T) {
	loan := createValidLoan()
	loan.State = model.StateApproved
	loan.NewInvestment = model.Investment{
		InvestorID: "investor-123",
		Name:       "John Doe",
		Email:      "john@example.com",
		Amount:     1500.0, // More than principal amount
	}

	sm := model.NewStateMachine(loan.State)
	err := sm.Transition(loan, model.EventAddInvestment)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "investment would exceed loan principal amount")
	assert.Equal(t, model.StateApproved, loan.State)
}

func TestInvalidStateTransition(t *testing.T) {
	loan := createValidLoan()
	loan.State = model.StateInitial

	sm := model.NewStateMachine(loan.State)
	err := sm.Transition(loan, model.EventDisburseFunds)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event disburse_funds not allowed in state initial")
}
