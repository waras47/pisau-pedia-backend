package mailer

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/config"
)

type Mailer struct {
	cfg config.SMTPConfig
}

func New(cfg config.SMTPConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) Enabled() bool {
	return m.cfg.Host != "" && m.cfg.Username != ""
}

func (m *Mailer) SendVerificationEmail(to, fullName, verifyURL string) error {
	subject := "Verifikasi Email Anda — Pisau Pedia"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px;background:#fafafa;">
  <div style="text-align:center;padding:20px 0;border-bottom:2px solid #1a1a2e;">
    <h1 style="color:#1a1a2e;margin:0;">Pisau Pedia</h1>
  </div>
  <div style="padding:30px 0;">
    <h2 style="color:#333;">Halo %s,</h2>
    <p style="color:#555;line-height:1.6;">
      Terima kasih telah mendaftar di Pisau Pedia! Silakan verifikasi email Anda dengan mengklik tombol di bawah ini:
    </p>
    <div style="text-align:center;padding:20px 0;">
      <a href="%s" style="background-color:#1a1a2e;color:#ffffff;padding:14px 32px;text-decoration:none;font-weight:bold;display:inline-block;border-radius:4px;">
        Verifikasi Email
      </a>
    </div>
    <p style="color:#555;line-height:1.6;">
      Atau salin link berikut ke browser Anda:<br>
      <a href="%s" style="color:#1a1a2e;word-break:break-all;">%s</a>
    </p>
    <p style="color:#999;font-size:13px;">
      Link ini berlaku selama 24 jam. Jika Anda tidak mendaftar di Pisau Pedia, abaikan email ini.
    </p>
  </div>
  <div style="border-top:1px solid #eee;padding-top:15px;text-align:center;color:#999;font-size:12px;">
    &copy; 2026 Pisau Pedia. Semua hak dilindungi.
  </div>
</body>
</html>`, fullName, verifyURL, verifyURL, verifyURL)

	msg := strings.Join([]string{
		fmt.Sprintf("From: Pisau Pedia <%s>", m.cfg.From),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	return smtp.SendMail(addr, auth, m.cfg.From, []string{to}, []byte(msg))
}

func (m *Mailer) SendOrderStatusEmail(to, customerName, orderID, oldStatus, newStatus string) error {
	statusLabels := map[string]string{
		"pending":            "Menunggu",
		"processing":         "Sedang Diproses",
		"ready_for_delivery": "Siap Dikirim",
		"delivered":          "Terkirim",
		"cancelled":          "Dibatalkan",
	}

	newLabel := statusLabels[newStatus]
	if newLabel == "" {
		newLabel = newStatus
	}

	badgeColor := "#16a34a"
	switch newStatus {
	case "pending":
		badgeColor = "#d97706"
	case "processing":
		badgeColor = "#2563eb"
	case "ready_for_delivery":
		badgeColor = "#7c3aed"
	case "cancelled":
		badgeColor = "#dc2626"
	}

	shortID := orderID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	subject := fmt.Sprintf("Update Pesanan #%s — Pisau Pedia", shortID)
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px;background:#fafafa;">
  <div style="text-align:center;padding:20px 0;border-bottom:2px solid #1a1a2e;">
    <h1 style="color:#1a1a2e;margin:0;">Pisau Pedia</h1>
  </div>
  <div style="padding:30px 0;">
    <h2 style="color:#333;">Halo %s,</h2>
    <p style="color:#555;line-height:1.6;">
      Status pesanan <strong>#%s</strong> Anda telah diperbarui:
    </p>
    <div style="text-align:center;padding:20px 0;">
      <span style="background-color:%s;color:#ffffff;padding:10px 24px;font-weight:bold;display:inline-block;border-radius:4px;">
        %s
      </span>
    </div>
    <p style="color:#555;line-height:1.6;">
      Anda dapat melihat detail pesanan di halaman <strong>Pesanan Saya</strong> pada akun Anda.
    </p>
    <p style="color:#555;line-height:1.6;">
      Jika ada pertanyaan, silakan balas email ini atau hubungi kami.
    </p>
  </div>
  <div style="border-top:1px solid #eee;padding-top:15px;text-align:center;color:#999;font-size:12px;">
    &copy; 2026 Pisau Pedia. Semua hak dilindungi.
  </div>
</body>
</html>`, customerName, shortID, badgeColor, newLabel)

	msg := strings.Join([]string{
		fmt.Sprintf("From: Pisau Pedia <%s>", m.cfg.From),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	if err := smtp.SendMail(addr, auth, m.cfg.From, []string{to}, []byte(msg)); err != nil {
		return err
	}
	return nil
}

func (m *Mailer) SendServiceRequestStatusEmail(to, customerName, requestType, oldStatus, newStatus, adminNotes string) error {
	statusLabels := map[string]string{
		"pending":     "Menunggu",
		"in_progress": "Sedang Diproses",
		"completed":   "Selesai",
		"rejected":    "Ditolak",
	}
	typeLabels := map[string]string{
		"sharpening": "Pengasahan Pisau",
		"engraving":  "Engraving Pisau",
	}

	typeLabel := typeLabels[requestType]
	if typeLabel == "" {
		typeLabel = requestType
	}
	newLabel := statusLabels[newStatus]
	if newLabel == "" {
		newLabel = newStatus
	}

	notesSection := ""
	if adminNotes != "" {
		notesSection = fmt.Sprintf(`
    <div style="background:#f5f5f5;padding:15px;border-radius:4px;margin:15px 0;">
      <p style="color:#333;font-weight:bold;margin:0 0 5px 0;">Catatan dari tim kami:</p>
      <p style="color:#555;margin:0;">%s</p>
    </div>`, adminNotes)
	}

	subject := fmt.Sprintf("Update Status %s — Pisau Pedia", typeLabel)
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px;background:#fafafa;">
  <div style="text-align:center;padding:20px 0;border-bottom:2px solid #1a1a2e;">
    <h1 style="color:#1a1a2e;margin:0;">Pisau Pedia</h1>
  </div>
  <div style="padding:30px 0;">
    <h2 style="color:#333;">Halo %s,</h2>
    <p style="color:#555;line-height:1.6;">
      Status permintaan <strong>%s</strong> Anda telah diperbarui:
    </p>
    <div style="text-align:center;padding:20px 0;">
      <span style="background-color:#16a34a;color:#ffffff;padding:10px 24px;font-weight:bold;display:inline-block;border-radius:4px;">
        %s
      </span>
    </div>
    %s
    <p style="color:#555;line-height:1.6;">
      Jika ada pertanyaan, silakan balas email ini atau hubungi kami.
    </p>
  </div>
  <div style="border-top:1px solid #eee;padding-top:15px;text-align:center;color:#999;font-size:12px;">
    &copy; 2026 Pisau Pedia. Semua hak dilindungi.
  </div>
</body>
</html>`, customerName, typeLabel, newLabel, notesSection)

	msg := strings.Join([]string{
		fmt.Sprintf("From: Pisau Pedia <%s>", m.cfg.From),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	return smtp.SendMail(addr, auth, m.cfg.From, []string{to}, []byte(msg))
}
