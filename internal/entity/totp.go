package entity

import (
	"crypto/rand"
	"errors"
	"regexp"
	"strings"
)

const recoveryAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

const (
	TOTPIssuer          = "Opsybot"
	TOTPDigits          = 6
	TOTPPeriodSeconds   = 30
	RecoveryCodeCount   = 8
	RecoveryCodeGroups  = 2
	RecoveryGroupLength = 4
)

var (
	totpCodeRe     = regexp.MustCompile(`^\d{6}$`)
	recoveryCodeRe = regexp.MustCompile(`^[a-z0-9]{4}-[a-z0-9]{4}$`)
)

var (
	ErrTOTPNotEnrolled  = errors.New("totp not enrolled")
	ErrTOTPAlreadySetUp = errors.New("totp already enabled")
	ErrTOTPInvalidCode  = errors.New("totp invalid code")
	ErrTOTPUnavailable  = errors.New("totp unavailable")
	ErrTOTPRequired     = errors.New("totp required")
	ErrRecoveryInvalid  = errors.New("recovery code invalid")
	ErrPendingNotFound  = errors.New("pending two-factor not found")
)

type TOTPEnrollment struct {
	Secret     string
	OTPAuthURI string
}

func ValidTOTPCode(code string) bool {
	return totpCodeRe.MatchString(code)
}

func ValidRecoveryCode(code string) bool {
	return recoveryCodeRe.MatchString(code)
}

func NormalizeRecoveryCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}

func GenerateRecoveryCodes(count int) ([]string, error) {
	length := RecoveryCodeGroups * RecoveryGroupLength
	out := make([]string, 0, count)
	for range count {
		buf := make([]byte, length)
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		var sb strings.Builder
		for i, b := range buf {
			if i > 0 && i%RecoveryGroupLength == 0 {
				sb.WriteByte('-')
			}
			sb.WriteByte(recoveryAlphabet[int(b)%len(recoveryAlphabet)])
		}
		out = append(out, sb.String())
	}
	return out, nil
}
