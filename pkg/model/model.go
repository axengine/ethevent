package model

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

type Task struct {
	ID        uint   `db:"id"`
	Contract  string `db:"contract"`
	Abi       string `db:"abi"`
	ChainId   uint64 `db:"chainId"`
	Rpc       string `db:"rpc"`
	Interval  int64  `db:"interval"` // 区块轮询间隔
	Start     uint64 `db:"start"`
	Current   uint64 `db:"current"`
	Paused    uint   `db:"paused"` // 是否暂停
	DeletedAt int64  `db:"deletedAt"`
	UpdatedAt int64  `db:"updatedAt"`
}

func (t *Task) TablePrefix() string {
	return fmt.Sprintf("EVENT_%d_", t.ID)
}

func (t *Task) TableName(eventName string) string {
	return fmt.Sprintf("EVENT_%d_%s", t.ID, eventName)
}

func (t *Task) TableNames() map[string]string {
	ins, _ := abi.JSON(strings.NewReader(t.Abi))
	var tables = make(map[string]string)
	for _, v := range ins.Events {
		tables[v.Name] = fmt.Sprintf("event_%d_%s", t.ID, v.Name)
	}

	return tables
}

type EventBase struct {
	ID          int64  `json:"id"`
	Address     string `json:"address"`
	BlockNumber uint64 `json:"blockNumber"`
	BlockHash   string `json:"blockHash"`
	BlockTime   int64  `json:"blockTime"`
	TxHash      string `json:"txHash"`
	TxIndex     uint   `json:"txIndex"`
	Method      uint32 `json:"method"`
}
