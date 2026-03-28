package tests

import (
	"testing"

	"github.com/C5rogers/G-Synch/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestGetColoredMessage_Warning(t *testing.T) {
	c := models.CheckReturn{Message: "test warning", Type: "MISMATCH", Label: "WARNING"}
	result := c.GetColoredMessage()
	assert.Contains(t, result, "test warning")
}

func TestGetColoredMessage_Error(t *testing.T) {
	c := models.CheckReturn{Message: "test error", Type: "ERROR", Label: "ERROR"}
	result := c.GetColoredMessage()
	assert.Contains(t, result, "test error")
}

func TestGetColoredMessage_Info(t *testing.T) {
	c := models.CheckReturn{Message: "test info", Type: "INFO", Label: "INFO"}
	result := c.GetColoredMessage()
	assert.Contains(t, result, "test info")
}

func TestGetColoredMessage_Success(t *testing.T) {
	c := models.CheckReturn{Message: "test success", Type: "OK", Label: "SUCCESS"}
	result := c.GetColoredMessage()
	assert.Contains(t, result, "test success")
}

func TestGetColoredMessage_Dependency(t *testing.T) {
	c := models.CheckReturn{Message: "dependency issue", Type: "MISSING_DEPENDENCY", Label: "DEPENDENCY"}
	result := c.GetColoredMessage()
	assert.Contains(t, result, "dependency issue")
}

func TestGetColoredMessage_Unknown(t *testing.T) {
	c := models.CheckReturn{Message: "unknown label", Type: "OTHER", Label: "UNKNOWN"}
	result := c.GetColoredMessage()
	assert.Equal(t, "unknown label", result)
}

func TestCMDMapper_ContainsAllCommands(t *testing.T) {
	expectedCmds := map[models.CMD]string{
		models.CHECK:         "check",
		models.SYNCH:         "synch",
		models.REVERSE_CHECK: "reverse-check",
	}

	for cmd, expected := range expectedCmds {
		actual, exists := models.CMDMapper[cmd]
		assert.True(t, exists, "CMDMapper should contain %s", cmd)
		assert.Equal(t, expected, actual)
	}
}

func TestCMDMapper_Length(t *testing.T) {
	assert.Len(t, models.CMDMapper, 3)
}
