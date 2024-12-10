package service

import (
	"context"
	"loan-engine/model"
	"loan-engine/repository"

	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockLoanRepository struct {
	mock.Mock
}

func (m *MockLoanRepository) Create(ctx context.Context, loan *model.Loan) (string, error) {
	args := m.Called(ctx, loan)
	return args.String(0), args.Error(1)
}

func (m *MockLoanRepository) Update(ctx context.Context, loan *model.Loan) error {
	args := m.Called(ctx, loan)
	return args.Error(0)
}

func (m *MockLoanRepository) GetLoan(ctx context.Context, id string) (*model.Loan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Loan), args.Error(1)
}

func (m *MockLoanRepository) CreateTransition(ctx context.Context, transition *model.Transition) error {
	args := m.Called(ctx, transition)
	return args.Error(0)
}

func (m *MockLoanRepository) CreateInvestment(ctx context.Context, loan *model.Loan) error {
	args := m.Called(ctx, loan)
	return args.Error(0)
}

func (m *MockLoanRepository) GetInvestments(ctx context.Context, loanID string) ([]model.Investment, error) {
	args := m.Called(ctx, loanID)
	return args.Get(0).([]model.Investment), args.Error(1)
}

func (m *MockLoanRepository) WithTransaction(ctx context.Context, fn func(repo repository.LoanRepositoryInterface) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// Mock Email Service
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendInvestmentAgreement(ctx context.Context, agreementURL string, loan *model.Loan) error {
	args := m.Called(ctx, agreementURL, loan)
	return args.Error(0)
}
