package tests

// test the config.go Load function to load yml file and read and return map[string]string
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/C5rogers/G-Synch/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// create a temporary config file
	tmpFile, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	configContent := `
key1: value1
key2: value2
key3: value3
`
	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Errorf("failed to write config file: %v", err)
	}
	tmpFile.Close()

	// make the file path absolute
	absPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		t.Errorf("failed to get absolute path: %v", err)
	}

	// load the config
	configs, err := config.Load(absPath)
	if err != nil {
		t.Errorf("failed to load config: %v", err)
	}

	// assert the values
	expectedConfigs := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	assert.Equal(t, expectedConfigs, configs)
}
