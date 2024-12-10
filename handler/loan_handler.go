package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"loan-engine/model"
	"loan-engine/service"

	"github.com/go-chi/chi/v5"
)

type LoanHandler struct {
	service *service.LoanService
}

func NewLoanHandler(service *service.LoanService) *LoanHandler {
	return &LoanHandler{service: service}
}

func (h *LoanHandler) CreateLoan(w http.ResponseWriter, r *http.Request) {
	var req model.CreateLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	loan, err := h.service.CreateLoan(r.Context(), req)
	if err != nil {
		JSONErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONSuccessResponse(w, http.StatusCreated, "Loan created successfully", loan.ID)
}

func (h *LoanHandler) ApproveLoan(w http.ResponseWriter, r *http.Request) {
	loanID := chi.URLParam(r, "id")
	if loanID == "" {
		JSONErrorResponse(w, http.StatusBadRequest, "loan id is required")
		return
	}

	var req model.ApproveLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	req.LoanID = loanID

	err := h.service.ApproveLoan(r.Context(), req)
	if err != nil {
		JSONErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONSuccessResponse(w, http.StatusOK, "Loan approved successfully", "")
}

func (h *LoanHandler) AddInvestment(w http.ResponseWriter, r *http.Request) {
	loanID := chi.URLParam(r, "id")
	if loanID == "" {
		JSONErrorResponse(w, http.StatusBadRequest, "loan id is required")
		return
	}

	var req model.AddInvestmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	req.LoanID = loanID

	isInvested, err := h.service.AddInvestment(r.Context(), req)
	if err != nil {
		JSONErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	data := ""
	if isInvested {
		data = fmt.Sprintf("Loan %s is already Invested", loanID)
	}

	JSONSuccessResponse(w, http.StatusCreated, "Loan investment created successfully", data)
}

func (h *LoanHandler) DisburseLoan(w http.ResponseWriter, r *http.Request) {
	loanID := chi.URLParam(r, "id")
	if loanID == "" {
		JSONErrorResponse(w, http.StatusBadRequest, "loan id is required")
		return
	}

	var req model.DisburseLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	req.LoanID = loanID

	err := h.service.DisburseLoan(r.Context(), req)
	if err != nil {
		JSONErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONSuccessResponse(w, http.StatusOK, "Loan disbursed successfully", "")
}

type SuccessResponse struct {
	Status  string      `json:"status"`         // e.g., "success"
	Message string      `json:"message"`        // Optional, can explain the success
	Data    interface{} `json:"data,omitempty"` // Any data to return
}

type ErrorResponse struct {
	Status  string `json:"status"`  // e.g., "error"
	Message string `json:"message"` // Error message for the client
	Code    int    `json:"code"`    // Optional, HTTP status code
}

func JSONSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

func JSONErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := ErrorResponse{
		Status:  "error",
		Message: message,
		Code:    statusCode,
	}
	json.NewEncoder(w).Encode(response)
}
