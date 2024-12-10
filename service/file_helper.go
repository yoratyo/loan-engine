package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"loan-engine/model"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/jung-kurt/gofpdf"
)

// generate file
func (s *LoanService) GenerateLoanAgreement(l *model.Loan) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Loan Agreement")
	pdf.Ln(20)

	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(190, 10, fmt.Sprintf(`
This Loan Agreement is made between:

Borrower: %s
Loan ID: %s
Principal Amount: %.2f

By signing this agreement, the borrower agrees to repay the loan in accordance with the terms specified herein.
`, l.BorrowerID, l.ID, l.PrincipalAmount), "", "", false)

	// Save to file
	fileName := fmt.Sprintf("loan_agreement_%s.pdf", l.ID)
	err := pdf.OutputFileAndClose(fileName)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

func (s *LoanService) UploadToFileIO(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}
	writer.Close()

	// Send POST request to File.io
	req, err := http.NewRequest("POST", "https://file.io", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["success"].(bool) {
		// Delete the file after successful upload
		err = os.Remove(filePath)
		if err != nil {
			return "", fmt.Errorf("file uploaded, but failed to delete local file: %w", err)
		}

		return result["link"].(string), nil
	}
	return "", fmt.Errorf("upload failed: %v", result)
}

func (s *LoanService) GenerateAndUploadLoanAgreement(l *model.Loan) (string, error) {
	// Generate the PDF
	filePath, err := s.GenerateLoanAgreement(l)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Upload the file to File.io
	link, err := s.UploadToFileIO(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return link, nil
}

// validate file
