package ethapi

import (
	"context"
	"crypto/ecdsa"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type HeatBeatAPI struct {
	b Backend
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
