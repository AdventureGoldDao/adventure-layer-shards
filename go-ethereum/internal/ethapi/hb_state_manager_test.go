package ethapi

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	testPrivKey      = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
	testStateContent = `[{"address":"0x1234567890123456789012345678901234567890","interval":5000000000}]`
)

func setup(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "ethapi-test")
	assert.NoError(t, err)

	_ = os.Setenv("HEAT_BEAT_PRIVATE_KEY", testPrivKey)

	return tmpDir, func() {
		os.Unsetenv("HEAT_BEAT_PRIVATE_KEY")
		_ = os.RemoveAll(tmpDir)
	}
}

func TestNewStateManager(t *testing.T) {
	tmpDir, cleanup := setup(t)
	defer cleanup()

	stateFile := filepath.Join(tmpDir, "test-state.json")
	err := os.WriteFile(stateFile, []byte(testStateContent), 0644)
	assert.NoError(t, err)

	sm := StateManager{
		StateFile: stateFile,
	}
	err = sm.LoadAll()
	assert.NoError(t, err)

	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	task := sm.LoadOne(addr)
	assert.NotNil(t, task)
	assert.Equal(t, 5*time.Second, task.Interval)
}

func TestStateManagerLifecycle(t *testing.T) {
	_, cleanup := setup(t)
	defer cleanup()

	sm := NewStateManager()
	assert.NotNil(t, sm)

	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	task := &ContractTask{
		Address:  addr,
		Interval: 3 * time.Second,
	}

	sm.Save(task)
	loadedTask := sm.LoadOne(addr)
	assert.Equal(t, task.Interval, loadedTask.Interval)

	sm.Delete(addr)
	assert.Nil(t, sm.LoadOne(addr))

	file, err := os.Open(sm.StateFile)
	assert.NoError(t, err)
	defer file.Close()

	var states []StateData
	err = json.NewDecoder(file).Decode(&states)
	assert.NoError(t, err)
	assert.Empty(t, states)
}

func TestStateManagerConcurrency(t *testing.T) {
	_, cleanup := setup(t)
	defer cleanup()

	sm := NewStateManager()
	assert.NotNil(t, sm)

	const workers = 100
	done := make(chan struct{})

	for i := 0; i < workers; i++ {
		go func(index int) {
			addr := common.BigToAddress(big.NewInt(int64(index)))
			task := &ContractTask{
				Address:  addr,
				Interval: time.Duration(index) * time.Second,
			}
			sm.Save(task)
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < workers; i++ {
		<-done
	}

	count := 0
	sm.ContractMap.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, workers, count)
}

func TestStateManagerErrorHandling(t *testing.T) {
	t.Run("InvalidPrivateKey", func(t *testing.T) {
		_ = os.Setenv("HEAT_BEAT_PRIVATE_KEY", "invalid-key")
		defer os.Unsetenv("HEAT_BEAT_PRIVATE_KEY")

		sm := NewStateManager()
		assert.Nil(t, sm)
	})

	t.Run("CorruptedStateFile", func(t *testing.T) {
		tmpDir, cleanup := setup(t)
		defer cleanup()

		stateFile := filepath.Join(tmpDir, "corrupted.json")
		err := os.WriteFile(stateFile, []byte("{invalid json}"), 0644)
		assert.NoError(t, err)

		sm := StateManager{
			StateFile: stateFile,
		}
		err = sm.LoadAll()
		assert.Error(t, err)
	})
}

func TestFilePersistence(t *testing.T) {
	tmpDir, cleanup := setup(t)
	defer cleanup()

	sm1 := StateManager{
		StateFile: filepath.Join(tmpDir, "state.json"),
	}
	addr := common.HexToAddress("0x123")
	sm1.Save(&ContractTask{
		Address:  addr,
		Interval: 10 * time.Second,
	})

	sm2 := StateManager{
		StateFile: filepath.Join(tmpDir, "state.json"),
	}
	err := sm2.LoadAll()
	assert.NoError(t, err)

	task := sm2.LoadOne(addr)
	assert.Equal(t, 10*time.Second, task.Interval)
}
