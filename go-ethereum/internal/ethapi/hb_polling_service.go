package ethapi

import (
	"context"
	"github.com/ethereum/go-ethereum/log"
	"time"
)

func (hb *HeatBeatAPI) startPolling(ctx context.Context, task *ContractTask) {
	data, err := mustPackABI()
	if err != nil {
		log.Error("contractABI:", err)
		return
	}

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Polling stopped", "address", task.Address.Hex())
			return
		case <-ticker.C:
			if err := hb.sendHeatBeatTransaction(ctx, task, stateManager.PrivateKey, data); err != nil {
				log.Error("Transaction failed", "error", err)
				return
			}
		}
	}
}
