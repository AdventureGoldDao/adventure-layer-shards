package heatbeat

import (
	"context"
	"crypto/ecdsa"
	_ "crypto/ecdsa"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	_ "github.com/ethereum/go-ethereum/crypto"
)

type HeatBeatAPI struct {
	b ethapi.Backend
}

type ContractTask struct {
	SendTxMutex sync.Mutex
	CancelFunc  context.CancelFunc
	Interval    time.Duration
	Address     common.Address
}

type StateManager struct {
	StateFile   string
	PrivateKey  *ecdsa.PrivateKey
	ContractMap sync.Map
}

const (
	defaultGasMultiplier = 2
	defaultStateFilename = "contract_tasks.json"
)

var (
	stateManager *StateManager
)
