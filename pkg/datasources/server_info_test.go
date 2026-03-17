package datasources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnwrapObjectData(t *testing.T) {
	wrapped := map[string]interface{}{
		"data": map[string]interface{}{
			"serverName": "vbr01",
			"version":    "13.0.0",
		},
	}

	result := unwrapObjectData(wrapped)

	assert.Equal(t, "vbr01", getStringValue(result, "serverName"))
	assert.Equal(t, "13.0.0", getStringValue(result, "version"))
}

func TestGetFirstStringValue(t *testing.T) {
	payload := map[string]interface{}{
		"server_name":  "vbr01.local",
		"build_number": "13.0.1.1234",
	}

	serverName := getFirstStringValue(payload, "serverName", "server_name", "name")
	buildNumber := getFirstStringValue(payload, "buildNumber", "build_number", "build")
	version := getFirstStringValue(payload, "version", "productVersion", "apiVersion")

	assert.Equal(t, "vbr01.local", serverName)
	assert.Equal(t, "13.0.1.1234", buildNumber)
	assert.Equal(t, "", version)
}
