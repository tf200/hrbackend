package twofa

import (
	"encoding/base64"
	"math/rand"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	charset    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLength = 8
)

func GenerateOTPSecret(issuer, accountName string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
	if err != nil {
		return "", err
	}
	return key.Secret(), nil
}

func ValidateCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

func GenerateQRCode(data string) (string, error) {
	qr, err := qrcode.Encode(data, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(qr), nil
}

func GenerateRecoveryCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		codes[i] = randomRecoveryCode()
	}
	return codes
}

func randomRecoveryCode() string {
	b := make([]byte, codeLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
