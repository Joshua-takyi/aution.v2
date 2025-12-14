package helpers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"path/filepath"

	"github.com/resend/resend-go/v3"
)

func SendVerificationEmail(resendClient *resend.Client, email string, verificationLink string) error {
	// Load and parse the HTML template
	templatePath := filepath.Join("internal", "templates", "verify-email.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	// Prepare template data
	data := struct {
		VerificationLink string
	}{
		VerificationLink: verificationLink,
	}

	// Execute template to generate HTML
	var htmlBuffer bytes.Buffer
	if err := tmpl.Execute(&htmlBuffer, data); err != nil {
		return err
	}

	// Send email using SendEmailRequest (not Broadcast!)
	params := &resend.SendEmailRequest{
		From:    "Acme <onboarding@resend.dev>",
		To:      []string{email}, // Send to the actual user
		Subject: "Verify Your Email Address",
		Html:    htmlBuffer.String(),
	}

	_, err = resendClient.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}

func GenerateVerificationToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)

}
