package chainindex

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Indexer interface {
	ID() int
	Contract() common.Address
	ABI() abi.ABI
	TableName() string
	InsertSQL() string
}
