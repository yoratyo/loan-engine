-- schema.sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE loans (
    id UUID PRIMARY KEY,
    borrower_id VARCHAR(255) NOT NULL,
    principal_amount DECIMAL(15,2) NOT NULL,
    total_investment_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    rate DECIMAL(5,2) NOT NULL,
    roi DECIMAL(15,2) NOT NULL,
    state VARCHAR(50) NOT NULL,
    field_validator_id VARCHAR(50),
    proof_image_url TEXT,
    approval_date TIMESTAMP,
    field_officer_id VARCHAR(50),
    agreement_letter_url TEXT,
    signed_agreement_letter_url TEXT,
    disbursement_date TIMESTAMP,
    version INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_set_updated_at_loans
BEFORE UPDATE ON loans
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE loan_investments (
    id UUID PRIMARY KEY,
    loan_id UUID REFERENCES loans(id) NOT NULL,
    investor_id VARCHAR(50) NOT NULL,
    investor_name VARCHAR(50) NOT NULL,
    email VARCHAR(50) NOT NULL,
	amount     DECIMAL(15,2) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_set_updated_at_loan_investments
BEFORE UPDATE ON loan_investments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE loan_state_transitions (
    id SERIAL PRIMARY KEY,
    loan_id UUID NOT NULL,
    previous_state VARCHAR(50) NOT NULL,
    event VARCHAR(50) NOT NULL,
    next_state VARCHAR(50) NOT NULL,
    transition_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_loans_borrower_id ON loans(borrower_id);
CREATE INDEX idx_loans_state ON loans(state);
CREATE INDEX idx_loan_investments_loan_id ON loan_investments(loan_id);
CREATE INDEX idx_loan_investments_investor_id ON loan_investments(investor_id);