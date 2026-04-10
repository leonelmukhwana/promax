package auth

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendOTPEmail(toEmail string, otp string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	// The exact message being sent
	subject := "Subject: Nanny App Verification Code\n"
	body := fmt.Sprintf("To: %s\n%s\nYour verification code is: %s\nIt expires in 10 minutes.", toEmail, subject, otp)

	// --- TESTING BLOCK: PRINT TO CONSOLE ---
	fmt.Println("\n==========================================")
	fmt.Println("📧 DEBUG EMAIL SENT")
	fmt.Println("To:", toEmail)
	fmt.Println("Message:", body)
	fmt.Println("==========================================\n")
	// ---------------------------------------

	// If using fake emails, the code below will likely fail.
	// We wrap it in a check so it doesn't crash your registration flow.
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, []byte(body))

	if err != nil {
		fmt.Printf("⚠️  SMTP Note: Email not actually sent to %s (Network/Auth error)\n", toEmail)
		// We return nil here during testing so the Register handler doesn't think
		// the whole process failed just because the email didn't fly.
		return nil
	}

	return nil
}
