package ethapi

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func NewHeatBeatAPI(b Backend) *HeatBeatAPI {
	hb := &HeatBeatAPI{b}
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

func (hb *HeatBeatAPI) ManageContractTask(address string, interval int, start bool) string {
	if address == "" || interval < 0 {
		return fmt.Sprintf("params err!")
	}

	addr := common.HexToAddress(address)
	if start {
		return hb.startTask(addr, interval)
	}
	return hb.stopTask(addr)
}

func (hb *HeatBeatAPI) startTask(addr common.Address, interval int) string {
	if task := stateManager.LoadOne(addr); task != nil {
		task.CancelFunc()
	}

	ctx, cancel := context.WithCancel(context.Background())
	task := &ContractTask{
		CancelFunc: cancel,
		Interval:   time.Duration(interval) * time.Millisecond,
		Address:    addr,
	}

	go hb.startPolling(ctx, task)

	stateManager.Save(task)
	return fmt.Sprintf("Started polling for contract: %s", addr.Hex())
}

func (hb *HeatBeatAPI) stopTask(addr common.Address) string {
	task := stateManager.LoadOne(addr)
	if task != nil {
		task.CancelFunc()
		stateManager.Delete(addr)
	}
	return fmt.Sprintf("Stop polling for contract: %s", addr.Hex())
}
