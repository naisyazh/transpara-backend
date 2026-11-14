package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/transpara/backend/pkg/config"
)

// EmailService menangani pengiriman email melalui Gmail SMTP.
type EmailService struct {
	cfg *config.Config
}

// NewEmailService membuat instance EmailService baru.
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

// SendOTP mengirim email berisi kode OTP ke alamat tujuan.
// Email dikirim menggunakan STARTTLS pada port 587.
//
// Parameter:
//   - toEmail : alamat email penerima (email yang dimasukkan user saat register)
//   - toName  : nama penerima untuk sapaan di body email
//   - otpCode : kode OTP 6 digit yang sudah di-generate
func (s *EmailService) SendOTP(toEmail, toName, otpCode string) error {
	// ── 1. Susun konten email (HTML) ─────────────────────────────
	body, err := buildOTPEmailBody(toName, otpCode, s.cfg.AppName)
	if err != nil {
		return fmt.Errorf("gagal membuat template email: %w", err)
	}

	// ── 2. Susun header email ─────────────────────────────────────
	from := fmt.Sprintf("%s <%s>", s.cfg.SMTPSenderName, s.cfg.SMTPUsername)
	subject := fmt.Sprintf("Kode Verifikasi OTP - %s", s.cfg.AppName)

	// Format pesan email lengkap dengan header MIME
	message := buildEmailMessage(from, toEmail, subject, body)

	// ── 3. Autentikasi ke Gmail SMTP ──────────────────────────────
	// Gmail menggunakan PlainAuth dengan App Password
	auth := smtp.PlainAuth(
		"",
		s.cfg.SMTPUsername,
		s.cfg.SMTPPassword,
		s.cfg.SMTPHost,
	)

	// ── 4. Kirim email via STARTTLS (port 587) ────────────────────
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	err = smtp.SendMail(
		addr,
		auth,
		s.cfg.SMTPUsername,
		[]string{toEmail},
		[]byte(message),
	)
	if err != nil {
		return fmt.Errorf("gagal mengirim email: %w", err)
	}

	return nil
}

// buildOTPEmailBody membuat body email HTML menggunakan template.
func buildOTPEmailBody(name, otpCode, appName string) (string, error) {
	const tmpl = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: Arial, sans-serif; background: #f4f4f4; margin: 0; padding: 20px; }
    .container { max-width: 480px; margin: 0 auto; background: #ffffff; border-radius: 8px; padding: 32px; }
    .header { text-align: center; margin-bottom: 24px; }
    .header h2 { color: #2563eb; margin: 0; }
    .otp-box { background: #eff6ff; border: 2px dashed #2563eb; border-radius: 8px;
                text-align: center; padding: 20px; margin: 24px 0; }
    .otp-code { font-size: 36px; font-weight: bold; letter-spacing: 8px; color: #1d4ed8; }
    .info { color: #6b7280; font-size: 14px; line-height: 1.6; }
    .warning { color: #dc2626; font-size: 13px; margin-top: 16px; }
    .footer { text-align: center; color: #9ca3af; font-size: 12px; margin-top: 24px; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h2>{{ .AppName }}</h2>
      <p style="color:#6b7280; margin:4px 0 0;">Verifikasi Akun Anda</p>
    </div>

    <p class="info">Halo <strong>{{ .Name }}</strong>,</p>
    <p class="info">Gunakan kode OTP berikut untuk memverifikasi akun Anda:</p>

    <div class="otp-box">
      <div class="otp-code">{{ .OTPCode }}</div>
    </div>

    <p class="info">Kode ini berlaku selama <strong>5 menit</strong> dan hanya bisa digunakan satu kali.</p>
    <p class="warning">⚠️ Jangan bagikan kode ini kepada siapapun, termasuk tim {{ .AppName }}.</p>

    <div class="footer">
      <p>Email ini dikirim otomatis oleh sistem {{ .AppName }}.<br>Abaikan jika Anda tidak merasa mendaftar.</p>
    </div>
  </div>
</body>
</html>`

	t, err := template.New("otp").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]string{
		"Name":    name,
		"OTPCode": otpCode,
		"AppName": appName,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// buildEmailMessage menyusun pesan email lengkap dengan header MIME.
func buildEmailMessage(from, to, subject, htmlBody string) string {
	return fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		from, to, subject, htmlBody,
	)
}
