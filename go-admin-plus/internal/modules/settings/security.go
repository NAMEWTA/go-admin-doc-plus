package settings

import (
	"regexp"
	"strings"
	"unicode"
)

var forbiddenMaterial = []string{
	"runtimeprofile", "databaseurl", "datasourcename", "postgresql", "postgres://", "mysql://", "jdbc:",
	"password", "passwd", "accesstoken", "refreshtoken", "sessiontoken", "privatekey", "secretkey", "clientsecret",
	"sessionpolicy", "idletimeout", "absolutetimeout", "rotationseconds", "beginprivatekey", "beginecprivatekey",
}

var (
	jwtMaterial       = regexp.MustCompile(`(?i)(?:^|[^a-z0-9_-])[a-z0-9_-]{8,}\.[a-z0-9_-]{8,}\.[a-z0-9_-]{8,}(?:$|[^a-z0-9_-])`)
	bearerMaterial    = regexp.MustCompile(`(?i)(?:^|\s)bearer\s+[a-z0-9._~+/-]{12,}`)
	userinfoMaterial  = regexp.MustCompile(`(?i)^[a-z][a-z0-9+.-]*://[^/@\s:]+:[^/@\s]+@`)
	highEntropySecret = regexp.MustCompile(`^[A-Za-z0-9_-]{43,}$`)
)

func sensitive(value string) bool {
	trimmed := strings.TrimSpace(value)
	canonical := strings.Map(func(current rune) rune {
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			return unicode.ToLower(current)
		}
		return -1
	}, trimmed)
	for _, marker := range forbiddenMaterial {
		marker = strings.Map(func(current rune) rune {
			if unicode.IsLetter(current) || unicode.IsDigit(current) {
				return unicode.ToLower(current)
			}
			return -1
		}, marker)
		if strings.Contains(canonical, marker) {
			return true
		}
	}
	return jwtMaterial.MatchString(trimmed) || bearerMaterial.MatchString(trimmed) || userinfoMaterial.MatchString(trimmed) || highEntropySecret.MatchString(trimmed)
}
