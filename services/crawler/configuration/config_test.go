package configuration

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestInitConfigSetsAValidLogger(t *testing.T) {

}
