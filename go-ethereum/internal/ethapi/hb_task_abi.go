package ethapi

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

var abiJSON = `[{"inputs":[],"name":"adventureHeatbeat","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

func mustPackABI() ([]byte, error) {
	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}
	return contractABI.Pack("adventureHeatbeat")
}
