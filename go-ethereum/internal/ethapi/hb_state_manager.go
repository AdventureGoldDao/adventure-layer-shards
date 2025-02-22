package ethapi

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func NewStateManager() *StateManager {
	ecdsa, err := crypto.HexToECDSA(os.Getenv("HEAT_BEAT_PRIVATE_KEY"))
	if err != nil {
		log.Error("Failed to load HEAT_BEAT_PRIVATE_KEY:", err)
		return nil
	}
	s := &StateManager{
		StateDir:   defaultStateDirname,
		PrivateKey: ecdsa,
	}
	err = s.LoadAll()
	if err != nil {
		log.Error("Failed to load NewStateManager:", err)
	}
	return s
}

func (sm *StateManager) LimitStatus() bool {
	hbLimitNum := os.Getenv("HEAT_BEAT_LIMIT_NUM")
	count := 0
	stateManager.ContractMap.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	hbLimitNumInt, err := strconv.Atoi(hbLimitNum)
	if err != nil {
		log.Error("Failed to convert HEAT_BEAT_LIMIT_NUM to int:", err)
		return true
	}
	return count >= hbLimitNumInt
}
func (sm *StateManager) Save(state *ContractTask) {
	sm.ContractMap.Store(state.AccountPublicKey.Hex(), state)
	sm.saveFile(state)
}

func (sm *StateManager) LoadOne(address common.Address) *ContractTask {
	if s, ok := sm.ContractMap.Load(address.Hex()); ok {
		return s.(*ContractTask)
	}
	return nil
}
func (sm *StateManager) LoadAll() error {
	if _, err := os.Stat(sm.StateDir); os.IsNotExist(err) {
		if err := os.MkdirAll(sm.StateDir, 0755); err != nil {
			log.Error("Failed to create state directory", "error", err)
			return err
		}
	}
	entries, err := os.ReadDir(sm.StateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || len(entry.Name()) != 42 {
			continue
		}
		filePath := filepath.Join(sm.StateDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Error("Failed to read state file", "file", filePath, "error", err)
			continue
		}

		var state StateData
		if err := json.Unmarshal(data, &state); err != nil {
			log.Error("Failed to parse state file", "file", filePath, "error", err)
			continue
		}
		taskState := new(ContractTask)
		taskState.AccountPublicKey = common.HexToAddress(state.AccountPublicKey)
		taskState.ContractAddress = common.HexToAddress(state.ContractAddress)
		taskState.Interval = state.Interval
		sm.ContractMap.Store(state.AccountPublicKey, taskState)
	}
	return nil
}

func (sm *StateManager) Delete(address common.Address) {
	sm.ContractMap.Delete(address.Hex())
	file := filepath.Join(sm.StateDir, address.Hex())
	log.Info("heartbeat deleteFile", "file", file)
	err := os.Remove(file)
	if err != nil {
		log.Error("Error deleting file: %v\n", err)
	}
}

func (sm *StateManager) saveFile(state *ContractTask) {
	fileName := filepath.Join(sm.StateDir, state.AccountPublicKey.Hex())
	file, err := os.Create(fileName)
	if err != nil {
		log.Error("Error creating file: %v\n", err)
		return
	}
	defer file.Close()
	data := StateData{
		AccountPublicKey: state.AccountPublicKey.Hex(),
		ContractAddress:  state.ContractAddress.Hex(),
		Interval:         state.Interval,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("Error marshaling JSON: %v\n", err)
		return
	}
	_, err = io.WriteString(file, string(jsonData))
	if err != nil {
		log.Error("Error writing to file: %v\n", err)
		return
	}
	log.Info("heartbeat saveFile success", "file", fileName)
}
