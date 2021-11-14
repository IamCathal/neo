package configuration

import (
	"os"
	"testing"

	"github.com/iamcathal/neo/services/crawler/datastructures"
	"gotest.tools/assert"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestInitAndSetWorkerConfig(t *testing.T) {
	os.Setenv("WORKER_AMOUNT", "8")
	expectedWorkerConfig := datastructures.WorkerConfig{
		WorkerAmount: 8,
	}
	InitAndSetWorkerConfig()

	assert.Equal(t, expectedWorkerConfig, WorkerConfig)
}
