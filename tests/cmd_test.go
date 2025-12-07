package tests

import (
	"testing"

	"github.com/C5rogers/G-Synch/internal/models"
)

func TestCMDModelToHaveTheDefinedTypedCommands(t *testing.T) {
	definedCommands := []string{"synch", "check", "reverse-check"}
	for _, cmd := range definedCommands {
		found := false
		for _, v := range models.CMDMapper {
			if v == cmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Command %s not found in CMDMapper", cmd)
		}
	}
}
