package notification

import (
	"context"
	"fmt"
	"loan-engine/config"
	"loan-engine/model"
	"log"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridService struct {
	client *sendgrid.Client
}

func NewSendGridService(apiKey string) *SendGridService {
	return &SendGridService{
		client: sendgrid.NewSendClient(apiKey),
	}
}

// SendBroadcastEmail sends a broadcast email to multiple recipients
func (s *SendGridService) SendInvestmentAgreement(ctx context.Context, agreementURL string, loan *model.Loan) error {
	cfg := config.LoadConfig()
	from := mail.NewEmail(cfg.EmailSenderName, cfg.EmailSenderAddress)

	// Create the personalizations (one for each recipient)
	var personalizations []*mail.Personalization
	for _, recipient := range loan.Investments {
		to := mail.NewEmail(recipient.Name, recipient.Email)
		p := mail.NewPersonalization()
		p.AddTos(to)
		personalizations = append(personalizations, p)
	}

	// Create the email message
	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = "Loan Investment Agreement"

	for _, p := range personalizations {
		message.AddPersonalizations(p)
	}

	// Add content to the email
	htmlContent := fmt.Sprintf(`
        <h2>Loan Investment Agreement</h2>
        <p>Dear Investor,</p>
        <p>Your loan investment has been fully funded. Please find your agreement letter at:</p>
        <p><a href="%s">View Agreement</a></p>
        <p>Best regards,<br>Loan Service Team</p>
    `, agreementURL)
	htmlTextContent := mail.NewContent("text/html", htmlContent)
	message.AddContent(htmlTextContent)

	// Send the email using SendGrid's client
	response, err := s.client.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email for loan ID [%s]: %w", loan.ID, err)
	}

	log.Printf("Email agreement for Loan ID [%s] sent with status code: %d, body: %s", loan.ID, response.StatusCode, response.Body)
	return nil
}
