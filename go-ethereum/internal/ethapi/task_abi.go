package ethapi

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

var abiJSON = `[{"inputs":[],"name":"adventureHeatbeat","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

func mustParseABI() (abi.ABI, error) {
	return abi.JSON(strings.NewReader(abiJSON))
}

func mustPackABI(abi abi.ABI) ([]byte, error) {
	return abi.Pack("adventureHeatbeat")
}
