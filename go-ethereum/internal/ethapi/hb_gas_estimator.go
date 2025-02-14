package ethapi

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
)

// GasPrice returns a suggestion for a gas price for legacy transactions.
func (hb *HeatBeatAPI) gasPrice(ctx context.Context) (*hexutil.Big, error) {
	tipCap, err := hb.b.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}
	if head := hb.b.CurrentHeader(); head.BaseFee != nil {
		tipCap.Add(tipCap, head.BaseFee)
	}
	return (*hexutil.Big)(tipCap), err
}

func (hb *HeatBeatAPI) estimateGas(ctx context.Context, fromAddr *common.Address, to *common.Address, data []byte) (uint64, error) {
	args := TransactionArgs{
		From:  fromAddr,
		To:    to,
		Data:  (*hexutil.Bytes)(&data),
		Value: (*hexutil.Big)(big.NewInt(0)),
	}
	blockNrOrHash := rpc.BlockNumberOrHashWithNumber(rpc.PendingBlockNumber)
	res, err := DoEstimateGas(ctx, hb.b, args, blockNrOrHash, nil, hb.b.RPCGasCap())
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %v", err)
	}
	return uint64(res), nil
}
