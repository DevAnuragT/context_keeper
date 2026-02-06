package services

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/DevAnuragT/context_keeper/internal/config"
)

// EmailService handles email operations
type EmailService interface {
	SendEmailVerification(ctx context.Context, email, token, firstName string) error
	SendPasswordReset(ctx context.Context, email, token, firstName string) error
	SendWelcomeEmail(ctx context.Context, email, firstName string) error
}

// EmailServiceImpl implements the EmailService interface
type EmailServiceImpl struct {
	config *config.Config
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.Config) EmailService {
	return &EmailServiceImpl{
		config: cfg,
	}
}

// SendEmailVerification sends an email verification email
func (e *EmailServiceImpl) SendEmailVerification(ctx context.Context, email, token, firstName string) error {
	if e.config.Email.SMTPHost == "" {
		// Email not configured, skip sending (for development)
		return nil
	}

	name := firstName
	if name == "" {
		name = "there"
	}

	subject := "Verify your email address"
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", e.config.Email.BaseURL, token)
	
	body := fmt.Sprintf(`
Hi %s,

Welcome to Context Keeper! Please verify your email address by clicking the link below:

%s

This link will expire in 24 hours.

If you didn't create an account with us, please ignore this email.

Best regards,
The Context Keeper Team
`, name, verificationURL)

	return e.sendEmail(email, subject, body)
}

// SendPasswordReset sends a password reset email
func (e *EmailServiceImpl) SendPasswordReset(ctx context.Context, email, token, firstName string) error {
	if e.config.Email.SMTPHost == "" {
		// Email not configured, skip sending (for development)
		return nil
	}

	name := firstName
	if name == "" {
		name = "there"
	}

	subject := "Reset your password"
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", e.config.Email.BaseURL, token)
	
	body := fmt.Sprintf(`
Hi %s,

You requested to reset your password for your Context Keeper account. Click the link below to set a new password:

%s

This link will expire in 1 hour.

If you didn't request a password reset, please ignore this email.

Best regards,
The Context Keeper Team
`, name, resetURL)

	return e.sendEmail(email, subject, body)
}

// SendWelcomeEmail sends a welcome email to new users
func (e *EmailServiceImpl) SendWelcomeEmail(ctx context.Context, email, firstName string) error {
	if e.config.Email.SMTPHost == "" {
		// Email not configured, skip sending (for development)
		return nil
	}

	name := firstName
	if name == "" {
		name = "there"
	}

	subject := "Welcome to Context Keeper!"
	
	body := fmt.Sprintf(`
Hi %s,

Welcome to Context Keeper! We're excited to have you on board.

Context Keeper helps you maintain context across your development projects by intelligently aggregating and organizing information from GitHub, Slack, Discord, and other platforms.

To get started:
1. Connect your GitHub repositories
2. Optionally connect Slack or Discord
3. Start querying your project context using our MCP tools

If you have any questions, feel free to reach out to our support team.

Best regards,
The Context Keeper Team
`, name)

	return e.sendEmail(email, subject, body)
}

// sendEmail sends an email using SMTP
func (e *EmailServiceImpl) sendEmail(to, subject, body string) error {
	// Create SMTP authentication
	auth := smtp.PlainAuth("", e.config.Email.SMTPUsername, e.config.Email.SMTPPassword, e.config.Email.SMTPHost)

	// Compose message
	msg := fmt.Sprintf("To: %s\r\nFrom: %s <%s>\r\nSubject: %s\r\n\r\n%s",
		to,
		e.config.Email.FromName,
		e.config.Email.FromAddress,
		subject,
		body)

	// Send email
	addr := fmt.Sprintf("%s:%d", e.config.Email.SMTPHost, e.config.Email.SMTPPort)
	err := smtp.SendMail(addr, auth, e.config.Email.FromAddress, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// MockEmailService is a mock implementation for testing
type MockEmailService struct {
	SentEmails []MockEmail
}

type MockEmail struct {
	To      string
	Subject string
	Body    string
	Type    string
}

// NewMockEmailService creates a new mock email service
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SentEmails: make([]MockEmail, 0),
	}
}

func (m *MockEmailService) SendEmailVerification(ctx context.Context, email, token, firstName string) error {
	m.SentEmails = append(m.SentEmails, MockEmail{
		To:      email,
		Subject: "Verify your email address",
		Body:    fmt.Sprintf("Verification token: %s", token),
		Type:    "verification",
	})
	return nil
}

func (m *MockEmailService) SendPasswordReset(ctx context.Context, email, token, firstName string) error {
	m.SentEmails = append(m.SentEmails, MockEmail{
		To:      email,
		Subject: "Reset your password",
		Body:    fmt.Sprintf("Reset token: %s", token),
		Type:    "password_reset",
	})
	return nil
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, email, firstName string) error {
	m.SentEmails = append(m.SentEmails, MockEmail{
		To:      email,
		Subject: "Welcome to Context Keeper!",
		Body:    fmt.Sprintf("Welcome %s!", firstName),
		Type:    "welcome",
	})
	return nil
}

// GetLastEmail returns the last sent email
func (m *MockEmailService) GetLastEmail() *MockEmail {
	if len(m.SentEmails) == 0 {
		return nil
	}
	return &m.SentEmails[len(m.SentEmails)-1]
}

// GetEmailsByType returns emails of a specific type
func (m *MockEmailService) GetEmailsByType(emailType string) []MockEmail {
	var emails []MockEmail
	for _, email := range m.SentEmails {
		if email.Type == emailType {
			emails = append(emails, email)
		}
	}
	return emails
}

// Clear clears all sent emails
func (m *MockEmailService) Clear() {
	m.SentEmails = make([]MockEmail, 0)
}