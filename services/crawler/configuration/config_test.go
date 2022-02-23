package configuration

import (
	"os"
	"sync"
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
	var waitG sync.WaitGroup
	waitG.Add(1)
	InitAndSetWorkerConfig(&waitG)
	waitG.Wait()
	assert.Equal(t, expectedWorkerConfig, WorkerConfig)
}
