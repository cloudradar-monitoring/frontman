package frontman

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func helperCreateFrontman(t *testing.T, cfg *Config) *Frontman {
	t.Helper()
	fm, err := New(cfg, DefaultCfgPath, "1.2.3")
	assert.Nil(t, err)
	fm.ipc = newIPC()
	return fm
}

func TestFrontmanHubInput(t *testing.T) {
	hub := NewMockHub("localhost:9100")
	go hub.Serve()

	cfg, err := HandleAllConfigSetup(DefaultCfgPath)
	assert.Nil(t, err)

	cfg.HubURL = hub.URL() + "/?serviceChecks=1&webChecks=1"
	cfg.LogLevel = "debug"
	cfg.Sleep = 10          // delay between each round of checks
	cfg.SenderBatchSize = 2 // number of results to send to hub at once
	cfg.SenderInterval = 0
	cfg.ICMPTimeout = 0.1
	cfg.HTTPCheckTimeout = 0.1
	cfg.SleepDurationAfterCheck = 0
	cfg.SleepDurationEmptyQueue = 0
	cfg.Nodes = make(map[string]Node)

	fm := helperCreateFrontman(t, cfg)

	go fm.Run("", nil)

	// stop after some time
	time.Sleep(2000 * time.Millisecond)
	close(fm.InterruptChan)

	fm.statsLock.Lock()
	assert.Equal(t, true, fm.stats.BytesSentToHubTotal > 0)
	assert.Equal(t, true, fm.stats.BytesFetchedFromHubTotal > 0)
	assert.Equal(t, true, fm.stats.ChecksPerformedTotal > 0)
	assert.Equal(t, uint64(2), fm.stats.ChecksFetchedFromHub)
	assert.Equal(t, true, fm.stats.CheckResultsSentToHub > 0)
	fm.statsLock.Unlock()

	fm.ipc.mutex.RLock()
	assert.Equal(t, true, len(fm.ipc.uuids) > 0)
	fm.ipc.mutex.RUnlock()
}
