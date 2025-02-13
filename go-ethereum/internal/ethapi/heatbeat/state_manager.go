package heatbeat

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"os"
	"path/filepath"
)

func NewStateManager() *StateManager {
	cacheDir, _ := os.UserCacheDir()
	ecdsa, err := crypto.HexToECDSA(os.Getenv("HEAT_BEAT_PRIVATE_KEY"))
	if err != nil {
		log.Error("Failed to load HEAT_BEAT_PRIVATE_KEY:", err)
		return nil
	}
	s := &StateManager{
		StateFile:  filepath.Join(cacheDir, defaultStateFilename),
		PrivateKey: ecdsa,
	}
	err = s.LoadAll()
	if err != nil {
		log.Error("Failed to load NewStateManager:", err)
	}
	return s
}

func (sm *StateManager) Save(state *ContractTask) error {
	sm.ContractMap.Store(state.Address, state)
	return nil
}

func (sm *StateManager) LoadOne(address common.Address) *ContractTask {
	if s, ok := sm.ContractMap.Load(address); ok {
		return s.(*ContractTask)
	}
	return nil
}
func (sm *StateManager) LoadAll() error {
	file, err := os.Open(sm.StateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var tasks []*ContractTask
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&tasks)
	if err != nil {
		return err
	}
	for _, taskState := range tasks {
		sm.ContractMap.Store(taskState.Address, taskState)
	}
	return nil
}

func (sm *StateManager) Delete(address common.Address) {
	sm.ContractMap.Delete(address)
}
