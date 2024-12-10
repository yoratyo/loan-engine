package model

import (
	"errors"
	"fmt"
)

type LoanState string

const (
	StateInitial   LoanState = "initial"
	StateProposed  LoanState = "proposed"
	StateApproved  LoanState = "approved"
	StateInvested  LoanState = "invested"
	StateDisbursed LoanState = "disbursed"
)

type LoanEvent string

const (
	EventSubmission    LoanEvent = "submission"
	EventApprove       LoanEvent = "approve"
	EventAddInvestment LoanEvent = "add_investment"
	EventDisburseFunds LoanEvent = "disburse_funds"
)

// Rule defines a function type for eligibility checks.
type EventRule func(l *Loan) (LoanState, error)

// Rule for submission event
func SubmissionRule(l *Loan) (LoanState, error) {
	if l.BorrowerID == "" {
		return l.State, errors.New("loan borrower ID data is empty")
	}

	if l.PrincipalAmount == 0 {
		return l.State, errors.New("loan principal amount data is empty")
	}

	if l.Rate == 0 {
		return l.State, errors.New("loan rate data is empty")
	}

	if l.ROI == 0 {
		return l.State, errors.New("loan roi data is empty")
	}

	return StateProposed, nil
}

// Rule for approving event
func ApproveRule(l *Loan) (LoanState, error) {
	if isStringNullOrEmpty(l.Approval.ProofImageURL) {
		return l.State, errors.New("approval proof image is empty")
	}

	if isStringNullOrEmpty(l.Approval.FieldValidatorID) {
		return l.State, errors.New("approval field validator is empty")
	}

	if isTimeNullOrEmpty(l.Approval.ApprovalDate) {
		return l.State, errors.New("approval date is empty")
	}

	return StateApproved, nil
}

// Rule for add investment event
func AddInvestmentRule(l *Loan) (LoanState, error) {
	// validate require fields
	if l.NewInvestment.InvestorID == "" {
		return l.State, errors.New("investor id for investment is empty")
	}

	if l.NewInvestment.Name == "" {
		return l.State, errors.New("investor name for investment is empty")
	}

	if l.NewInvestment.Email == "" {
		return l.State, errors.New("investor email for investment is empty")
	}

	if l.NewInvestment.Amount == 0 {
		return l.State, errors.New("investor amount for investment is empty")
	}

	// validate investment amount
	currentTotal := l.TotalInvestmentAmount + l.NewInvestment.Amount
	if currentTotal > l.PrincipalAmount {
		return l.State, errors.New("investment would exceed loan principal amount")
	}

	if currentTotal == l.PrincipalAmount {
		return StateInvested, nil
	}

	return StateApproved, nil
}

// Rule for disburse funds event
func DisburseFundsRule(l *Loan) (LoanState, error) {
	if isStringNullOrEmpty(l.Disbursement.SignedAgreementLetterURL) {
		return l.State, errors.New("disbursement agreement letter is empty")
	}

	if isStringNullOrEmpty(l.Disbursement.FieldOfficerID) {
		return l.State, errors.New("disbursement field officer is empty")
	}

	if isTimeNullOrEmpty(l.Disbursement.DisbursementDate) {
		return l.State, errors.New("disbursement date is empty")
	}

	return StateDisbursed, nil
}

// StateMachine represents a state machine.
type StateMachine struct {
	currentState LoanState
	transitions  map[LoanState]map[LoanEvent]EventRule // current state -> event -> EventRule(next state & validator)
}

// NewStateMachine initializes a new state machine.
func NewStateMachine(initialState LoanState) *StateMachine {
	return &StateMachine{
		currentState: initialState,
		transitions: map[LoanState]map[LoanEvent]EventRule{
			StateInitial: {
				EventSubmission: SubmissionRule,
			},
			StateProposed: {
				EventApprove: ApproveRule,
			},
			StateApproved: {
				EventAddInvestment: AddInvestmentRule,
			},
			StateInvested: {
				EventDisburseFunds: DisburseFundsRule,
			},
		},
	}
}

// Transition attempts to move the state machine to the next state.
func (sm *StateMachine) Transition(loan *Loan, event LoanEvent) error {
	allowedTransitions, ok := sm.transitions[sm.currentState]
	if !ok {
		return fmt.Errorf("invalid current state: %s", sm.currentState)
	}

	eventRule, ok := allowedTransitions[event]
	if !ok {
		return fmt.Errorf("event %s not allowed in state %s", event, sm.currentState)
	}

	// Determine the next state using the decision rule
	nextState, err := eventRule(loan)
	if err != nil {
		return fmt.Errorf("state decision failed: %v", err)
	}

	sm.currentState = nextState
	loan.State = nextState
	return nil
}

// GetCurrentState returns the current state.
func (sm *StateMachine) GetCurrentState() LoanState {
	return sm.currentState
}
