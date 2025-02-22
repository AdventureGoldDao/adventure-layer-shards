package ethapi

import (
	"context"
	"github.com/ethereum/go-ethereum/log"
	"time"
)

func (hb *HeartBeatAPI) startPolling(ctx context.Context, task *ContractTask) {
	abiData, err := mustPackABI()
	if err != nil {
		log.Error("contractABI:", err)
		return
	}

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Polling stopped", "address", task.ContractAddress.Hex())
			return
		case <-ticker.C:
			if err := hb.sendHeartBeatTransaction(ctx, task, stateManager.PrivateKey, abiData); err != nil {
				log.Error("Transaction failed", "error", err)
				return
			}
		}
	}
}
