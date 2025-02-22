package ethapi

import (
	"context"
	"crypto/ecdsa"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type HeartBeatAPI struct {
	b Backend
}

type ContractTask struct {
	SendTxMutex      sync.Mutex
	CancelFunc       context.CancelFunc
	Interval         time.Duration
	AccountPublicKey common.Address
	ContractAddress  common.Address
}

type StateData struct {
	Interval         time.Duration `json:"interval"`
	AccountPublicKey string        `json:"AccountPublicKey"`
	ContractAddress  string        `json:"contractAddress"`
}

type StateManager struct {
	StateDir    string
	PrivateKey  *ecdsa.PrivateKey
	ContractMap sync.Map
}

const (
	defaultGasMultiplier = 2
	defaultStateDirname  = "/config/heartbeat_lists"
)

var (
	stateManager *StateManager
)
