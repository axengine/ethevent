package model

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

type Task struct {
	ID       uint   `db:"id"`
	Contract string `db:"contract"`
	Abi      string `db:"abi"`
	ChainId  uint64 `db:"chainId"`
	Rpc      string `db:"rpc"`
	Begin    uint64 `db:"begin"`
	Current  uint64 `db:"current"`
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
	Address     string   `gorm:"address"`
	Topics      []string `gorm:"topics"`
	Data        []byte   `gorm:"data"`
	BlockNumber uint64   `gorm:"blockNumber"`
	TxHash      string   `gorm:"txHash"`
	TxIndex     uint     `gorm:"txIndex"`
	BlockHash   string   `gorm:"blockHash"`
	LogIndex    uint     `gorm:"logIndex"`
	Removed     bool     `gorm:"removed"`
}
