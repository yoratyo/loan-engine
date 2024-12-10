package service_test

import (
	"context"
	"loan-engine/model"
	"loan-engine/service"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function to create test loan
func createTestLoan() *model.Loan {
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

func TestCreateLoan(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(service.MockLoanRepository)
	mockEmail := new(service.MockEmailService)
	service := service.NewLoanService(mockRepo, mockEmail)

	testCases := []struct {
		name        string
		request     model.CreateLoanRequest
		setupMocks  func()
		expectError bool
	}{
		{
			name: "Successful loan creation",
			request: model.CreateLoanRequest{
				BorrowerID:      "borrower-123",
				PrincipalAmount: 1000.0,
				Rate:            5.0,
				ROI:             10.0,
			},
			setupMocks: func() {
				mockRepo.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Loan")).Return("loan-123", nil)
				mockRepo.On("CreateTransition", mock.Anything, mock.AnythingOfType("*model.Transition")).Return(nil)
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			loan, err := service.CreateLoan(ctx, tc.request)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, loan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, loan)
				assert.Equal(t, model.StateProposed, loan.State)
			}
		})
	}
}

func TestApproveLoan(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(service.MockLoanRepository)
	mockEmail := new(service.MockEmailService)
	service := service.NewLoanService(mockRepo, mockEmail)

	testCases := []struct {
		name        string
		request     model.ApproveLoanRequest
		setupMocks  func()
		expectError bool
	}{
		{
			name: "Successful loan approval",
			request: model.ApproveLoanRequest{
				LoanID:        "loan-123",
				ValidatorID:   "validator-123",
				ProofImageURL: "http://example.com/proof.jpg",
				ApprovalDate:  time.Now(),
			},
			setupMocks: func() {
				loan := createTestLoan()
				loan.State = model.StateProposed
				mockRepo.On("GetLoan", mock.Anything, "loan-123").Return(loan, nil)
				mockRepo.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Loan")).Return(nil)
				mockRepo.On("CreateTransition", mock.Anything, mock.AnythingOfType("*model.Transition")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "Failed approval - Invalid state",
			request: model.ApproveLoanRequest{
				LoanID:        "loan-123",
				ValidatorID:   "validator-123",
				ProofImageURL: "http://example.com/proof.jpg",
				ApprovalDate:  time.Now(),
			},
			setupMocks: func() {
				loan := createTestLoan()
				loan.State = model.StateInitial // Wrong state
				mockRepo.On("GetLoan", mock.Anything, "loan-123").Return(loan, nil)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			err := service.ApproveLoan(ctx, tc.request)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddInvestment(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(service.MockLoanRepository)
	mockEmail := new(service.MockEmailService)
	service := service.NewLoanService(mockRepo, mockEmail)

	testCases := []struct {
		name           string
		request        model.AddInvestmentRequest
		setupMocks     func()
		expectError    bool
		expectInvested bool
	}{
		{
			name: "Successful partial investment",
			request: model.AddInvestmentRequest{
				LoanID:     "loan-123",
				InvestorID: "investor-123",
				Amount:     500.0,
				Name:       "John Doe",
				Email:      "john@example.com",
			},
			setupMocks: func() {
				loan := createTestLoan()
				loan.State = model.StateApproved
				mockRepo.On("GetLoan", mock.Anything, "loan-123").Return(loan, nil)
				mockRepo.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Loan")).Return(nil)
				mockRepo.On("CreateInvestment", mock.Anything, mock.AnythingOfType("*model.Loan")).Return(nil)
				mockRepo.On("CreateTransition", mock.Anything, mock.AnythingOfType("*model.Transition")).Return(nil)
			},
			expectError:    false,
			expectInvested: false,
		},
		{
			name: "Successful full investment",
			request: model.AddInvestmentRequest{
				LoanID:     "loan-123",
				InvestorID: "investor-123",
				Amount:     1000.0,
				Name:       "John Doe",
				Email:      "john@example.com",
			},
			setupMocks: func() {
				loan := createTestLoan()
				loan.State = model.StateApproved
				mockRepo.On("GetLoan", mock.Anything, "loan-123").Return(loan, nil)
				mockRepo.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Loan")).Return(nil)
				mockRepo.On("CreateInvestment", mock.Anything, mock.AnythingOfType("*model.Loan")).Return(nil)
				mockRepo.On("CreateTransition", mock.Anything, mock.AnythingOfType("*model.Transition")).Return(nil)
				mockRepo.On("GetInvestments", mock.Anything, "loan-123").Return([]model.Investment{}, nil)
				mockEmail.On("SendInvestmentAgreement", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectError:    false,
			expectInvested: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			invested, err := service.AddInvestment(ctx, tc.request)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectInvested, invested)
			}
		})
	}
}

func TestDisburseLoan(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(service.MockLoanRepository)
	mockEmail := new(service.MockEmailService)
	service := service.NewLoanService(mockRepo, mockEmail)

	testCases := []struct {
		name        string
		request     model.DisburseLoanRequest
		setupMocks  func()
		expectError bool
	}{
		{
			name: "Successful loan disbursement",
			request: model.DisburseLoanRequest{
				LoanID:             "loan-123",
				OfficerID:          "validator-123",
				AgreementLetterURL: "http://example.com/proof.jpg",
				DisbursementDate:   time.Now(),
			},
			setupMocks: func() {
				loan := createTestLoan()
				loan.State = model.StateInvested
				mockRepo.On("GetLoan", mock.Anything, "loan-123").Return(loan, nil)
				mockRepo.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Loan")).Return(nil)
				mockRepo.On("CreateTransition", mock.Anything, mock.AnythingOfType("*model.Transition")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "Failed disbursement - Invalid state",
			request: model.DisburseLoanRequest{
				LoanID:             "loan-123",
				OfficerID:          "validator-123",
				AgreementLetterURL: "http://example.com/proof.jpg",
				DisbursementDate:   time.Now(),
			},
			setupMocks: func() {
				loan := createTestLoan()
				loan.State = model.StateInitial // Wrong state
				mockRepo.On("GetLoan", mock.Anything, "loan-123").Return(loan, nil)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			err := service.DisburseLoan(ctx, tc.request)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileOperations(t *testing.T) {
	service := &service.LoanService{}
	loan := createTestLoan()

	t.Run("Generate loan agreement", func(t *testing.T) {
		filePath, err := service.GenerateLoanAgreement(loan)
		assert.NoError(t, err)
		assert.FileExists(t, filePath)

		// Cleanup
		defer os.Remove(filePath)
	})
}
