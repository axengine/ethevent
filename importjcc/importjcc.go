package main

import (
	"context"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"github.com/axengine/ethcli"
	"github.com/axengine/ethevent/pkg/database"
	"github.com/axengine/utils/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"strings"
)

type Event struct {
	Table string
	Cols  []database.Feild
}

func main() {
	f, err := os.Open("goerli.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ',' // 设置分隔符为逗号

	cli, _ := ethcli.New("https://goerli.infura.io/v3/03d2548af36149abb66a54983ea238f9")
	defer cli.Close()
	ins, _ := abi.JSON(strings.NewReader(`[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`))

	dbo := database.New("events.db", log.Logger)

	handled := make(map[string]bool)
	for {
		records, err := r.Read()
		if err != nil {
			break
		}
		fmt.Println(records[0])
		if strings.ToUpper(records[0]) == "TXHASH" {
			continue
		}
		if handled[records[0]] {
			continue
		}
		handled[records[0]] = true

		var events []Event
		tx, _, err := cli.TransactionByHash(context.Background(), common.HexToHash(records[0]))
		if err != nil {
			panic(err)
		}
		receipt, err := cli.TransactionReceipt(context.Background(), common.HexToHash(records[0]))
		if err != nil {
			panic(err)
		}

		for _, rcptLog := range receipt.Logs {
			event, err := ins.EventByID(rcptLog.Topics[0])
			if err != nil {
				continue
			}

			eventAddress := rcptLog.Address.Hex()
			if eventAddress != common.HexToAddress("0xC01138c43c8D99732fa900059FCAA9f34Cd6047a").Hex() {
				continue
			}

			var cols []database.Feild
			{
				cols = append(cols, database.Feild{
					Name:  "Address",
					Value: eventAddress,
				})
				cols = append(cols, database.Feild{
					Name:  "BlockNumber",
					Value: rcptLog.BlockNumber,
				})
				cols = append(cols, database.Feild{
					Name:  "BlockHash",
					Value: rcptLog.BlockHash.Hex(),
				})
				cols = append(cols, database.Feild{
					Name:  "BlockTime",
					Value: records[2],
				})
				cols = append(cols, database.Feild{
					Name:  "TxHash",
					Value: rcptLog.TxHash.Hex(),
				})
				cols = append(cols, database.Feild{
					Name:  "TxIndex",
					Value: rcptLog.TxIndex,
				})
				cols = append(cols, database.Feild{
					Name:  "Method",
					Value: binary.BigEndian.Uint32(tx.Data()[:4]),
				})
			}

			// 索引参数和非索引参数在旧版本solidity中可以乱序
			var indexedParams = make(map[string]interface{})
			var indexedArgs = make([]abi.Argument, 0)
			for _, v := range event.Inputs {
				if v.Indexed {
					indexedArgs = append(indexedArgs, v)
				}
			}

			// 索引参数
			indexed := rcptLog.Topics[1:]
			if len(indexed) > 0 {
				if err := abi.ParseTopicsIntoMap(indexedParams, indexedArgs, indexed); err != nil {
					panic(err)
				}
				for k, v := range indexedParams {
					if vv, ok := v.(fmt.Stringer); ok {
						v = vv.String()
					}
					cols = append(cols, database.Feild{
						Name:  k,
						Value: v,
					})
				}
			}

			// 非索引参数
			if len(rcptLog.Data) > 0 {
				unindexed, err := event.Inputs.Unpack(rcptLog.Data)
				if err != nil {
					panic(err)
				}

				for i, v := range event.Inputs.NonIndexed() {
					if vv, ok := unindexed[i].(fmt.Stringer); ok {
						unindexed[i] = vv.String()
					}
					cols = append(cols, database.Feild{
						Name:  v.Name,
						Value: unindexed[i],
					})
				}
			}

			events = append(events, Event{
				Table: "EVENT_2_TRANSFER",
				Cols:  cols,
			})
		}

		for _, v := range events {
			if _, err := dbo.Insert(nil, v.Table, v.Cols); err != nil {
				panic(err)
			}
		}
	}
}
