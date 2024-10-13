package arbitrum

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"
)

// PublicNetAPI offers network related RPC methods
type PublicNetAPI struct {
	networkVersion uint64
}

// NewPublicNetAPI creates a new net API instance.
func NewPublicNetAPI(networkVersion uint64) *PublicNetAPI {
	return &PublicNetAPI{networkVersion}
}

// Version returns the current ethereum protocol version.
func (s *PublicNetAPI) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}

var contractMap sync.Map

type ContractTask struct {
	CancelFunc context.CancelFunc
	Interval   time.Duration
	PrivateKey string
	RpcUrl     string
	Address    common.Address
}

func (s *PublicNetAPI) ManageContractTask(address, privateKey, rpcUrl string, interval int, start bool) {
	addr := common.HexToAddress(address)
	if start {
		if _, exists := contractMap.Load(addr); exists {
			log.Printf("Polling task for contract %s is already running.", addr.Hex())
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		task := ContractTask{
			CancelFunc: cancel,
			Interval:   time.Duration(interval) * time.Millisecond,
			PrivateKey: privateKey,
			RpcUrl:     rpcUrl,
			Address:    addr,
		}

		contractMap.Store(addr, task)

		go startPolling(ctx, task)
		log.Printf("Started polling for contract: %s", addr.Hex())
	} else {
		if taskInterface, exists := contractMap.Load(addr); exists {
			task := taskInterface.(ContractTask)
			task.CancelFunc()
			contractMap.Delete(addr)
			log.Printf("Stopped polling for contract: %s", addr.Hex())
		} else {
			log.Printf("No polling task found for contract: %s", addr.Hex())
		}
	}
}

func startPolling(ctx context.Context, task ContractTask) {
	client, err := ethclient.Dial("http://localhost:8587") // 使用 Nitro 节点的 RPC 地址
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	key, err := crypto.HexToECDSA(task.PrivateKey)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	fromAddr := crypto.PubkeyToAddress(key.PublicKey)
	nonce, _ := client.PendingNonceAt(ctx, fromAddr)

	gasLimit := uint64(300000)
	gasPrice, _ := client.SuggestGasPrice(ctx)

	contractABI, err := abi.JSON(strings.NewReader(`[{"inputs":[],"name":"myFunction","outputs":[],"stateMutability":"nonpayable","type":"function"}]`))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Polling stopped for contract: %s", task.Address.Hex())
			return
		default:
			data, _ := contractABI.Pack("myFunction")
			txdata := &types.LegacyTx{
				Nonce:    nonce,
				To:       &task.Address,
				Value:    big.NewInt(0),
				Gas:      gasLimit,
				GasPrice: gasPrice,
				Data:     data,
			}
			tx := types.NewTx(txdata)
			signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, key)
			if err != nil {
				log.Printf("Failed to sign transaction: %v", err)
				continue
			}

			err = client.SendTransaction(ctx, signedTx)
			if err != nil {
				log.Printf("Failed to send transaction: %v", err)
			} else {
				log.Printf("Transaction sent: %s", signedTx.Hash().Hex())
				nonce++
			}

			time.Sleep(task.Interval)
		}
	}
}
