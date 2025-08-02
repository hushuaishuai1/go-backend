package services

import (
	"fmt"
	"log"
	"net/smtp"
	"github.com/hushuaishuai1/go-backend/internal/config"
	"github.com/hushuaishuai1/go-backend/internal/models"
)

type Mailer struct {
	cfg config.SMTPConfig
}

func NewMailer(cfg config.SMTPConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) SendAlertEmail(alert models.Alert, currentPrice float64) {
	to := []string{alert.Email}
	directionText := "高于"
	if alert.Direction == "BELOW" {
		directionText = "低于"
	}
	subject := fmt.Sprintf("价格警报: %s 已%s您的目标价!", alert.Symbol, directionText)
	body := fmt.Sprintf(
		"您好！\n您设置的价格警报已被触发。\n\n- 交易对: %s\n- 您的条件: 价格 %s %.4f\n- 当前价格: %.4f\n\n此警报已完成并被移除。",
		alert.Symbol, directionText, alert.TargetPrice, currentPrice,
	)

	msg := []byte("To: " + alert.Email + "\r\n" +
		"From: " + m.cfg.Sender + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body)

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	err := smtp.SendMail(addr, auth, m.cfg.Sender, to, msg)
	if err != nil {
		log.Printf("ERROR: Failed to send email to %s: %v", alert.Email, err)
	} else {
		log.Printf("SUCCESS: Sent email to %s for %s alert.", alert.Email, alert.Symbol)
	}
}