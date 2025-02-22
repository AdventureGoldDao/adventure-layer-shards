package ethapi

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func NewHeartBeatAPI(b Backend) *HeartBeatAPI {
	hb := &HeartBeatAPI{b}
	stateManager = NewStateManager()
	if stateManager != nil {
		stateManager.ContractMap.Range(func(key, value interface{}) bool {
			task := value.(*ContractTask)
			ctx, cancel := context.WithCancel(context.Background())
			task.CancelFunc = cancel
			go hb.startPolling(ctx, task)
			return true
		})
	}
	return hb
}

func (hb *HeartBeatAPI) ManageContractTask(contractAddress, accountPublicKey string, interval int, start bool) string {
	if contractAddress == "" || accountPublicKey == "" || interval < 0 {
		return fmt.Sprintf("params err!")
	}

	accountAddr := common.HexToAddress(accountPublicKey)
	contractAddr := common.HexToAddress(contractAddress)
	if start {
		return hb.startTask(accountAddr, contractAddr, interval)
	}
	return hb.stopTask(accountAddr)
}

func (hb *HeartBeatAPI) GetActiveHeartBeats() []map[string]string {
	var activeHeartBeats []map[string]string
	stateManager.ContractMap.Range(func(key, value interface{}) bool {
		task := value.(*ContractTask)
		activeHeartBeats = append(activeHeartBeats, map[string]string{
			"contractAddress":  task.ContractAddress.Hex(),
			"accountPublicKey": task.AccountPublicKey.Hex(),
		})
		return true
	})
	return activeHeartBeats
}

func (hb *HeartBeatAPI) startTask(accountPublicKey, contractAddress common.Address, interval int) string {
	if stateManager.LimitStatus() {
		return fmt.Sprintf("heartbeat task limit ...")
	}
	if task := stateManager.LoadOne(accountPublicKey); task != nil {
		task.CancelFunc()
	}
	ctx, cancel := context.WithCancel(context.Background())
	task := &ContractTask{
		CancelFunc:       cancel,
		Interval:         time.Duration(interval) * time.Millisecond,
		AccountPublicKey: accountPublicKey,
		ContractAddress:  contractAddress,
	}

	go hb.startPolling(ctx, task)

	stateManager.Save(task)
	return fmt.Sprintf("Started polling for contract: %s", contractAddress.Hex())
}

func (hb *HeartBeatAPI) stopTask(addr common.Address) string {
	task := stateManager.LoadOne(addr)
	if task == nil {
		return fmt.Sprintf("No active task found for accountAddr: %s", addr.Hex())
	}
	task.CancelFunc()
	stateManager.Delete(addr)
	log.Info("Stopped polling for contract: ", task.ContractAddress.Hex())
	return fmt.Sprintf("Stop polling for contract: %s", addr.Hex())
}
