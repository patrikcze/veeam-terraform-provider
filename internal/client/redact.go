package client

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

const redactedValue = "[REDACTED]"

var sensitiveAssignmentPattern = regexp.MustCompile(`(?i)\b(access[_-]?key|application[_-]?key|authorization|passphrase|password|private[_-]?key|refresh[_-]?token|secret|secret[_-]?key|shared[_-]?key|token)\b\s*[:=]\s*("[^"]*"|'[^']*'|[^\s,;)}\]]+)`)

func sanitizeAPIError(apiErr models.APIError) models.APIError {
	apiErr.ErrorCode = redactSensitiveText(apiErr.ErrorCode)
	apiErr.Message = redactSensitiveText(apiErr.Message)
	apiErr.Details = redactSensitiveText(apiErr.Details)
	return apiErr
}

func sanitizeErrorBody(body []byte, maxLen int) string {
	redactedBody := redactSensitiveJSON(body)
	if redactedBody == nil {
		redactedBody = []byte(redactSensitiveText(string(body)))
	}

	return truncateBody(redactedBody, maxLen)
}

func redactSensitiveJSON(body []byte) []byte {
	var value interface{}
	if err := json.Unmarshal(body, &value); err != nil {
		return nil
	}

	redactSensitiveJSONValue(value)

	redacted, err := json.Marshal(value)
	if err != nil {
		return nil
	}

	return bytes.TrimSpace(redacted)
}

func redactSensitiveJSONValue(value interface{}) {
	switch typedValue := value.(type) {
	case map[string]interface{}:
		for key, nestedValue := range typedValue {
			if isSensitiveFieldName(key) {
				typedValue[key] = redactedValue
				continue
			}
			redactSensitiveJSONValue(nestedValue)
		}
	case []interface{}:
		for _, nestedValue := range typedValue {
			redactSensitiveJSONValue(nestedValue)
		}
	}
}

func isSensitiveFieldName(name string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, "_", ""), "-", ""))
	for _, marker := range []string{
		"accesstoken",
		"refreshtoken",
		"authorization",
		"password",
		"passphrase",
		"privatekey",
		"secret",
		"secretkey",
		"sharedkey",
		"accesskey",
		"applicationkey",
		"tokenvalue",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}

	return normalized == "token" || normalized == "key"
}

func redactSensitiveText(value string) string {
	return sensitiveAssignmentPattern.ReplaceAllStringFunc(value, func(match string) string {
		separatorIndex := strings.IndexAny(match, ":=")
		if separatorIndex < 0 {
			return match
		}

		return match[:separatorIndex+1] + redactedValue
	})
}
