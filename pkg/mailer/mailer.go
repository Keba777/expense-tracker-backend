package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/rs/zerolog/log"
)

type Mailer struct {
	host     string
	port     int
	username string
	password string
	from     string
	enabled  bool
}

func New(host string, port int, username, password, from string) *Mailer {
	return &Mailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		enabled:  host != "",
	}
}

func (m *Mailer) SendPasswordReset(to, resetURL string) error {
	subject := "Reset your ExpenseTracker password"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:480px;margin:0 auto;padding:24px;color:#111">
  <h2 style="margin-bottom:8px">Reset your password</h2>
  <p style="color:#555;margin-bottom:24px">Click the button below to set a new password. This link expires in 30 minutes.</p>
  <a href="%s" style="display:inline-block;background:#7c3aed;color:#fff;text-decoration:none;padding:12px 24px;border-radius:8px;font-weight:600">
    Reset password
  </a>
  <p style="color:#888;font-size:13px;margin-top:24px">
    If you didn't request this, you can safely ignore this email.<br>
    Or copy this link: %s
  </p>
</body>
</html>`, resetURL, resetURL)

	if !m.enabled {
		log.Info().Str("to", to).Str("reset_url", resetURL).Msg("[dev] password reset link (SMTP not configured)")
		return nil
	}

	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		m.from, to, subject, body,
	))
	return smtp.SendMail(fmt.Sprintf("%s:%d", m.host, m.port), auth, m.from, []string{to}, msg)
}
