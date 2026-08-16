package mail

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/formsx/backend/internal/config"
)

// SendPlain sends a plain-text email using SMTP settings from cfg.
// Returns an error if SMTP is not configured or send fails.
func SendPlain(cfg *config.Config, to []string, subject, body string) error {
	if cfg.SMTPHost == "" || cfg.SMTPFrom == "" {
		return fmt.Errorf("SMTP not configured: set SMTP_HOST and SMTP_FROM")
	}
	port := cfg.SMTPPort
	if port == "" {
		port = "587"
	}
	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, port)

	var auth smtp.Auth
	if cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	}

	// Minimal RFC 5322 plain text message
	subject = strings.ReplaceAll(subject, "\r", " ")
	subject = strings.ReplaceAll(subject, "\n", " ")
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		cfg.SMTPFrom,
		strings.Join(to, ", "),
		subject,
		body,
	))

	return smtp.SendMail(addr, auth, cfg.SMTPFrom, to, msg)
}

// SendHTML sends an HTML email using SMTP settings from cfg.
func SendHTML(cfg *config.Config, to []string, subject, htmlBody string) error {
	if cfg.SMTPHost == "" || cfg.SMTPFrom == "" {
		return fmt.Errorf("SMTP not configured: set SMTP_HOST and SMTP_FROM")
	}
	port := cfg.SMTPPort
	if port == "" {
		port = "587"
	}
	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, port)

	var auth smtp.Auth
	if cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	}

	subject = strings.ReplaceAll(subject, "\r", " ")
	subject = strings.ReplaceAll(subject, "\n", " ")
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		cfg.SMTPFrom,
		strings.Join(to, ", "),
		subject,
		htmlBody,
	))

	return smtp.SendMail(addr, auth, cfg.SMTPFrom, to, msg)
}
