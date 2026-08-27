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

var reservedKeyNamespaces = map[string]struct{}{
	"auth": {}, "authentication": {}, "database": {}, "logging": {}, "observability": {},
	"runtime": {}, "security": {}, "server": {}, "session": {}, "telemetry": {},
}

var reservedKeyParts = map[string]struct{}{
	"credential": {}, "credentials": {}, "dsn": {}, "password": {}, "passwd": {},
	"secret": {}, "token": {},
}

var reservedCanonicalKeys = []string{
	"absolutetimeout", "accesskey", "apikey", "clientsecret", "connectionstring",
	"databaseurl", "datasourcename", "encryptionkey", "idletimeout", "listenaddress",
	"loglevel", "privatekey", "rotationseconds", "runtimeprofile", "secretkey",
	"sessionpolicy", "signingkey",
}

func sensitiveKey(value string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	parts := strings.FieldsFunc(trimmed, func(current rune) bool {
		return current == '.' || current == '-' || current == '_'
	})
	if len(parts) == 0 {
		return false
	}
	if _, reserved := reservedKeyNamespaces[parts[0]]; reserved {
		return true
	}
	for _, part := range parts {
		if _, reserved := reservedKeyParts[part]; reserved {
			return true
		}
	}
	canonical := canonicalMaterial(trimmed)
	for _, marker := range reservedCanonicalKeys {
		if strings.Contains(canonical, marker) {
			return true
		}
	}
	return false
}

func sensitive(value string) bool {
	trimmed := strings.TrimSpace(value)
	canonical := canonicalMaterial(trimmed)
	for _, marker := range forbiddenMaterial {
		if strings.Contains(canonical, canonicalMaterial(marker)) {
			return true
		}
	}
	return jwtMaterial.MatchString(trimmed) || bearerMaterial.MatchString(trimmed) || userinfoMaterial.MatchString(trimmed) || highEntropySecret.MatchString(trimmed)
}

func canonicalMaterial(value string) string {
	return strings.Map(func(current rune) rune {
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			return unicode.ToLower(current)
		}
		return -1
	}, value)
}
