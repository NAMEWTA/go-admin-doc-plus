package outbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

type PayloadKind uint8

const (
	PayloadString PayloadKind = iota + 1
	PayloadNumber
	PayloadBoolean
)

type PayloadFieldSchema struct {
	Name     string
	Kind     PayloadKind
	Required bool
}

type BusinessKeySchema struct {
	Prefix   string
	MinParts int
	MaxParts int
}

type TopicSchema struct {
	Topic       string
	Payload     []PayloadFieldSchema
	BusinessKey BusinessKeySchema
}

type topicValidator struct {
	payload     map[string]PayloadFieldSchema
	businessKey BusinessKeySchema
}

var (
	payloadFieldPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{0,63}$`)
	keyPartPattern      = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
)

func compileTopicSchemas(schemas []TopicSchema) (map[string]topicValidator, error) {
	if len(schemas) == 0 {
		return nil, errors.New("outbox topic schema is required")
	}
	compiled := make(map[string]topicValidator, len(schemas))
	for _, schema := range schemas {
		if !topicPattern.MatchString(schema.Topic) || schema.BusinessKey.Prefix == "" ||
			!keyPartPattern.MatchString(schema.BusinessKey.Prefix) ||
			schema.BusinessKey.MinParts < 1 || schema.BusinessKey.MaxParts < schema.BusinessKey.MinParts ||
			schema.BusinessKey.MaxParts > 8 || isSensitiveBusinessPart(schema.BusinessKey.Prefix) {
			return nil, errors.New("outbox topic schema is invalid")
		}
		if _, exists := compiled[schema.Topic]; exists {
			return nil, errors.New("outbox topic schema is duplicated")
		}
		fields := make(map[string]PayloadFieldSchema, len(schema.Payload))
		for _, field := range schema.Payload {
			if !payloadFieldPattern.MatchString(field.Name) || !validPayloadFieldName(field.Name) ||
				field.Kind < PayloadString || field.Kind > PayloadBoolean {
				return nil, errors.New("outbox payload schema is invalid")
			}
			if _, exists := fields[field.Name]; exists {
				return nil, errors.New("outbox payload schema is duplicated")
			}
			fields[field.Name] = field
		}
		compiled[schema.Topic] = topicValidator{payload: fields, businessKey: schema.BusinessKey}
	}
	return compiled, nil
}

func (validator topicValidator) validate(payload []byte, businessKey string) bool {
	parts := strings.Split(businessKey, ":")
	if len(parts) < 2 || parts[0] != validator.businessKey.Prefix {
		return false
	}
	parts = parts[1:]
	if len(parts) < validator.businessKey.MinParts || len(parts) > validator.businessKey.MaxParts {
		return false
	}
	for _, part := range parts {
		if !keyPartPattern.MatchString(part) || isSensitiveBusinessPart(part) {
			return false
		}
	}

	if !json.Valid(payload) {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil || object == nil {
		return false
	}
	for name, value := range object {
		field, allowed := validator.payload[name]
		if !allowed || !matchesPayloadKind(value, field.Kind) {
			return false
		}
	}
	for name, field := range validator.payload {
		if _, exists := object[name]; field.Required && !exists {
			return false
		}
	}
	return true
}

func matchesPayloadKind(value any, kind PayloadKind) bool {
	switch kind {
	case PayloadString:
		_, ok := value.(string)
		return ok
	case PayloadNumber:
		_, ok := value.(json.Number)
		return ok
	case PayloadBoolean:
		_, ok := value.(bool)
		return ok
	default:
		return false
	}
}

func validPayloadFieldName(name string) bool {
	tokens := identifierTokens(name)
	for index, token := range tokens {
		if !sensitiveToken(token) {
			continue
		}
		if index+1 == len(tokens)-1 && metadataToken(tokens[index+1]) {
			continue
		}
		return false
	}
	return true
}

func identifierTokens(value string) []string {
	var builder strings.Builder
	for index, character := range value {
		if index > 0 && character >= 'A' && character <= 'Z' {
			builder.WriteByte(' ')
		}
		if character == '_' || character == '-' || character == '.' {
			builder.WriteByte(' ')
			continue
		}
		builder.WriteRune(character)
	}
	return strings.Fields(strings.ToLower(builder.String()))
}

func sensitiveToken(value string) bool {
	switch value {
	case "password", "secret", "session", "token", "credential":
		return true
	default:
		return false
	}
}

func metadataToken(value string) bool {
	switch value {
	case "count", "timeout", "ttl", "enabled", "required", "status", "type", "expires", "expiry":
		return true
	default:
		return false
	}
}

func isSensitiveBusinessPart(part string) bool {
	lower := strings.ToLower(part)
	for _, token := range []string{"password", "secret", "session", "token", "credential"} {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}
