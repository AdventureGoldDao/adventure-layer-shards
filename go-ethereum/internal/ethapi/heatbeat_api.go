package ethapi

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func NewHeatBeatAPI(b Backend) *HeatBeatAPI {
	s := &HeatBeatAPI{b}
	stateManager = NewStateManager()
	stateManager.ContractMap.Range(func(key, value interface{}) bool {
		task := value.(*ContractTask)
		ctx, cancel := context.WithCancel(context.Background())
		task.CancelFunc = cancel
		go s.startPolling(ctx, task)
		return true
	})
	return s
}

func (tm *HeatBeatAPI) ManageContractTask(address string, interval int, start bool) string {
	if address == "" || interval < 0 {
		return fmt.Sprintf("params err!")
	}

	addr := common.HexToAddress(address)
	if start {
		return tm.startTask(addr, interval)
	}
	return tm.stopTask(addr)
}

func (tm *HeatBeatAPI) startTask(addr common.Address, interval int) string {
	if task := stateManager.LoadOne(addr); task != nil {
		task.CancelFunc()
	}

	ctx, cancel := context.WithCancel(context.Background())
	task := &ContractTask{
		CancelFunc: cancel,
		Interval:   time.Duration(interval) * time.Millisecond,
		Address:    addr,
	}

	go tm.startPolling(ctx, task)

	err := stateManager.Save(task)
	if err != nil {
		return fmt.Sprintf("stateManager Save error, %v", err)
	}
	return fmt.Sprintf("Started polling for contract: %s", addr.Hex())
}

func (tm *HeatBeatAPI) stopTask(addr common.Address) string {
	task := stateManager.LoadOne(addr)
	task.CancelFunc()
	stateManager.Delete(addr)
	return fmt.Sprintf("Stop polling for contract: %s", addr.Hex())
}

func (tm *HeatBeatAPI) startPolling(ctx context.Context, task *ContractTask) {

	contractABI, err := mustParseABI()
	if err != nil {
		log.Error("Failed to parse contract ABI:", err)
		return
	}
	data, err := mustPackABI(contractABI)
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
			if err := tm.sendTransaction(ctx, task, stateManager.PrivateKey, data); err != nil {
				log.Error("Transaction failed", "error", err)
				return
			}
		}
	}
}

func (tm *HeatBeatAPI) sendTransaction(ctx context.Context, task *ContractTask, key *ecdsa.PrivateKey, data []byte) error {
	task.SendTxMutex.Lock()
	defer task.SendTxMutex.Unlock()

	fromAddr := crypto.PubkeyToAddress(key.PublicKey)

	gasLimit, err := tm.estimateGas(ctx, &fromAddr, &task.Address, data)
	if err != nil {
		return fmt.Errorf("gas estimation failed: %w", err)
	}

	gasPrice, err := tm.gasPrice(ctx)
	if err != nil {
		return fmt.Errorf("gas price fetch failed: %w", err)
	}

	nonce, err := tm.b.GetPoolNonce(ctx, fromAddr)
	if err != nil {
		return fmt.Errorf("nonce fetch failed: %w", err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &task.Address,
		Value:    new(big.Int),
		Gas:      gasLimit,
		GasPrice: big.NewInt(gasPrice.ToInt().Int64() * defaultGasMultiplier),
		Data:     data,
	})

	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, key)
	if err != nil {
		return fmt.Errorf("tx signing failed: %w", err)
	}

	if _, err = SubmitTransaction(ctx, tm.b, signedTx); err != nil {
		return fmt.Errorf("tx submission failed: %w", err)
	}

	log.Info("Transaction sent", "hash", signedTx.Hash().Hex())
	return nil
}

// GasPrice returns a suggestion for a gas price for legacy transactions.
func (s *HeatBeatAPI) gasPrice(ctx context.Context) (*hexutil.Big, error) {
	tipCap, err := s.b.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}
	if head := s.b.CurrentHeader(); head.BaseFee != nil {
		tipCap.Add(tipCap, head.BaseFee)
	}
	return (*hexutil.Big)(tipCap), err
}

func (s *HeatBeatAPI) estimateGas(ctx context.Context, fromAddr *common.Address, to *common.Address, data []byte) (uint64, error) {
	args := TransactionArgs{
		From:  fromAddr,
		To:    to,
		Data:  (*hexutil.Bytes)(&data),
		Value: (*hexutil.Big)(big.NewInt(0)),
	}
	blockNrOrHash := rpc.BlockNumberOrHashWithNumber(rpc.PendingBlockNumber)
	res, err := DoEstimateGas(ctx, s.b, args, blockNrOrHash, nil, s.b.RPCGasCap())
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %v", err)
	}
	return uint64(res), nil
}
