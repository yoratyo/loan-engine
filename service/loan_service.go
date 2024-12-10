package service

import (
	"context"
	"log"

	"loan-engine/model"
	repo "loan-engine/repository"
)

type EmailService interface {
	SendInvestmentAgreement(ctx context.Context, agreementURL string, loan *model.Loan) error
}

type LoanService struct {
	repo  repo.LoanRepositoryInterface
	email EmailService
}

func NewLoanService(repo repo.LoanRepositoryInterface, email EmailService) *LoanService {
	return &LoanService{repo: repo, email: email}
}

func (s *LoanService) CreateLoan(ctx context.Context, r model.CreateLoanRequest) (*model.Loan, error) {
	var loanID string
	loan := &model.Loan{
		BorrowerID:      r.BorrowerID,
		PrincipalAmount: r.PrincipalAmount,
		Rate:            r.Rate,
		ROI:             r.ROI,
		State:           model.StateProposed,
	}
	// Initialize the state machine
	loanStateMachine := model.NewStateMachine(model.StateInitial)
	// Transition to "proposed"
	err := loanStateMachine.Transition(loan, model.EventSubmission)
	if err != nil {
		return nil, err
	}

	err = s.repo.WithTransaction(ctx, func(rTx repo.LoanRepositoryInterface) error {
		var err error
		loanID, err = rTx.Create(ctx, loan)
		if err != nil {
			return err
		}

		transition := &model.Transition{
			LoanID:        loanID,
			PreviousState: model.StateInitial,
			Event:         model.EventSubmission,
			NextState:     model.StateProposed,
		}

		err = rTx.CreateTransition(ctx, transition)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	loan.ID = loanID

	return loan, nil
}

func (s *LoanService) ApproveLoan(ctx context.Context, r model.ApproveLoanRequest) error {
	loan, err := s.repo.GetLoan(ctx, r.LoanID)
	if err != nil {
		return err
	}
	loan.Approval = r.ToApproval()

	previousState := loan.State
	// Initialize current the state machine
	loanStateMachine := model.NewStateMachine(previousState)
	// Transition to "approve"
	err = loanStateMachine.Transition(loan, model.EventApprove)
	if err != nil {
		return err
	}

	err = s.repo.WithTransaction(ctx, func(rTx repo.LoanRepositoryInterface) error {
		err := rTx.Update(ctx, loan)
		if err != nil {
			return err
		}

		transition := &model.Transition{
			LoanID:        loan.ID,
			PreviousState: previousState,
			Event:         model.EventApprove,
			NextState:     loanStateMachine.GetCurrentState(),
		}

		err = rTx.CreateTransition(ctx, transition)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *LoanService) AddInvestment(ctx context.Context, r model.AddInvestmentRequest) (bool, error) {
	loan, err := s.repo.GetLoan(ctx, r.LoanID)
	if err != nil {
		return false, err
	}
	loan.NewInvestment = r.ToInvestment()

	previousState := loan.State
	// Initialize current the state machine
	loanStateMachine := model.NewStateMachine(previousState)
	// Transition to "add_investment"
	err = loanStateMachine.Transition(loan, model.EventAddInvestment)
	if err != nil {
		return false, err
	}

	err = s.repo.WithTransaction(ctx, func(rTx repo.LoanRepositoryInterface) error {
		if loan.State == model.StateInvested {
			agreementURL, err := s.GenerateAndUploadLoanAgreement(loan)
			if err != nil {
				log.Printf("Failed to generate agreement for loan %s: %v", loan.ID, err)
				return err
			}
			// set link agreement url
			loan.SetAgreementURL(agreementURL)
		}
		err := rTx.Update(ctx, loan)
		if err != nil {
			return err
		}

		err = rTx.CreateInvestment(ctx, loan)
		if err != nil {
			return err
		}

		transition := &model.Transition{
			LoanID:        loan.ID,
			PreviousState: previousState,
			Event:         model.EventAddInvestment,
			NextState:     loanStateMachine.GetCurrentState(),
		}

		err = rTx.CreateTransition(ctx, transition)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	// send email if state invested
	if loanStateMachine.GetCurrentState() == model.StateInvested {
		// Async email sending
		go func(loan *model.Loan) {
			asyncCtx := context.Background()
			// Get list investor
			listInvestments, err := s.repo.GetInvestments(asyncCtx, loan.ID)
			if err != nil {
				log.Printf("Failed to get investments for loan %s: %v", loan.ID, err)
			}
			loan.Investments = listInvestments

			// Send broadcast email
			err = s.email.SendInvestmentAgreement(asyncCtx, loan.AgreementLetterURL.String, loan)
			if err != nil {
				log.Printf("Failed to send email for loan %s: %v", loan.ID, err)
			}
		}(loan)
		log.Printf("Send email for loan %s is IN PROGRESS", loan.ID)

		return true, nil
	}

	return false, nil
}

func (s *LoanService) DisburseLoan(ctx context.Context, r model.DisburseLoanRequest) error {
	loan, err := s.repo.GetLoan(ctx, r.LoanID)
	if err != nil {
		return err
	}
	loan.Disbursement = r.ToDisbursement()

	previousState := loan.State
	// Initialize current the state machine
	loanStateMachine := model.NewStateMachine(previousState)
	// Transition to "disburse_funds"
	err = loanStateMachine.Transition(loan, model.EventDisburseFunds)
	if err != nil {
		return err
	}

	err = s.repo.WithTransaction(ctx, func(rTx repo.LoanRepositoryInterface) error {
		err := rTx.Update(ctx, loan)
		if err != nil {
			return err
		}

		transition := &model.Transition{
			LoanID:        loan.ID,
			PreviousState: previousState,
			Event:         model.EventDisburseFunds,
			NextState:     loanStateMachine.GetCurrentState(),
		}

		err = rTx.CreateTransition(ctx, transition)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
