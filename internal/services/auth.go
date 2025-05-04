package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"bankapp/internal/config"
	"github.com/go-mail/mail/v2"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}
func CheckPasswordHash(pw, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// Генерация номера счета
func generateAccountNumber() string {
	base := "400000"
	for len(base) < 15 {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		base += fmt.Sprintf("%d", n.Int64())
	}
	sum := 0
	for i, c := range base {
		d := int(c - '0')
		if i%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	check := (10 - sum%10) % 10
	return base + fmt.Sprintf("%d", check)
}

// Генерация номера карты
func generateCardNumber() string {
	return generateAccountNumber()[:16]
}

// Срок действия (по умолчанию 3 года)
func generateExpiryDate() (month, year int) {
	t := time.Now().AddDate(3, 0, 0)
	return int(t.Month()), t.Year()
}

func generateCVV() string {
	cvv := ""
	for i := 0; i < 3; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		cvv += fmt.Sprintf("%d", n.Int64())
	}
	return cvv
}

// go-mail
func sendEmailNotification(cfg *config.Config, to, subject, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", cfg.SMTPUser)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := mail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	return d.DialAndSend(m)
}
