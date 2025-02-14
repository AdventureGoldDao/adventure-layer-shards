package ethapi

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"io"
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

func (sm *StateManager) Save(state *ContractTask) {
	sm.ContractMap.Store(state.Address, state)
	sm.saveFile()
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

	var states []StateData
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&states)
	if err != nil {
		return err
	}
	for _, state := range states {
		taskState := new(ContractTask)
		taskState.Address = common.HexToAddress(state.Address)
		taskState.Interval = state.Interval
		sm.ContractMap.Store(taskState.Address, taskState)
	}
	return nil
}

func (sm *StateManager) Delete(address common.Address) {
	sm.ContractMap.Delete(address)
	sm.saveFile()
}

func (sm *StateManager) saveFile() {
	file, err := os.Create(sm.StateFile)
	if err != nil {
		log.Error("Error creating file: %v\n", err)
		return
	}
	defer file.Close()
	var states []StateData
	sm.ContractMap.Range(func(key, value interface{}) bool {
		task := value.(*ContractTask)
		states = append(states, StateData{
			Address:  task.Address.String(),
			Interval: task.Interval,
		})
		return true
	})
	jsonData, err := json.Marshal(states)
	if err != nil {
		log.Error("Error marshaling JSON: %v\n", err)
		return
	}
	_, err = io.WriteString(file, string(jsonData))
	if err != nil {
		log.Error("Error writing to file: %v\n", err)
		return
	}
	log.Info("heatbeat data saveFile success")
}
