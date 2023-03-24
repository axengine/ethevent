package chainindex

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/axengine/ethcli"
	"github.com/axengine/ethevent/pkg/dbo"
	"github.com/axengine/ethevent/pkg/model"
	"github.com/axengine/utils/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"math/big"
	"reflect"
	"strings"
)

type ChainIndex struct {
	db *dbo.DBO
}

func New(db *dbo.DBO) *ChainIndex {
	return &ChainIndex{db: db}
}

// Init initial table and index
func (ci *ChainIndex) Init() error {
	// create table TASK
	if _, err := ci.db.Exec(nil, model.CreateTaskTableSQL); err != nil {
		log.Logger.Error("Exec", zap.Error(err), zap.String("sql", model.CreateTaskTableSQL))
		return err
	}

	var tasks []model.Task
	where := []dbo.Where{{Name: "1", Value: 1}}
	if err := ci.db.SelectRows("TASK", where, nil, nil, &tasks); err != nil {
		panic(err)
	}

	// create event table
	for _, v := range tasks {
		var tablePrefix = fmt.Sprintf("EVENT_%d_", v.ID)
		ins, err := abi.JSON(strings.NewReader(v.Abi))
		if err != nil {
			return err
		}
		for _, v := range ins.Events {
			tableName := tablePrefix + strings.ToUpper(v.Name)
			var createCols []string
			var indexCols []string
			for _, arg := range v.Inputs {
				var tp string
				switch arg.Type.T {
				case abi.BoolTy, abi.IntTy, abi.UintTy:
					tp = "INTEGER"
				default:
					tp = "TEXT"
				}
				createCols = append(createCols, fmt.Sprintf("%s %s", "["+strings.ToUpper(arg.Name)+"]", tp))
				if arg.Indexed {
					indexCols = append(indexCols, fmt.Sprintf(`%s`, strings.ToUpper(arg.Name)))
				}
			}

			ctsqls := model.CreateTableSQL(tableName, createCols)
			if _, err := ci.db.Exec(nil, ctsqls); err != nil {
				log.Logger.Error("Exec", zap.Error(err), zap.String("sql", ctsqls))
				return err
			}

			cisqls := model.CreateIndexSQL(tableName, indexCols)
			for _, v := range cisqls {
				if _, err := ci.db.Exec(nil, v); err != nil {
					log.Logger.Error("Exec", zap.Error(err), zap.String("sql", v))
					return err
				}
			}
		}
	}

	return nil
}

func (ci *ChainIndex) Start(ctx context.Context) error {
	var tasks []model.Task
	where := []dbo.Where{{Name: "1", Value: 1}}
	if err := ci.db.SelectRows("TASK", where, nil, nil, &tasks); err != nil {
		panic(err)
	}

	for _, v := range tasks {
		if cli, err := ethcli.New(v.Rpc); err != nil {
			return err
		} else {
			if err := ci.start(ctx, cli, &v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ci *ChainIndex) start(ctx context.Context, cli *ethcli.ETHCli, t *model.Task) error {
	for {
		select {
		case <-ctx.Done():
			log.Logger.Info("ChainIndex exit")
			return nil
		default:
			number, err := cli.BlockNumber(ctx)
			if err != nil {
				log.Logger.Warn("BlockNumber", zap.Error(err))
				continue
			}
			if t.Current < number {
				if err := ci.handleNumber(ctx, cli, t.Current+1, t); err != nil {
					log.Logger.Error("handleNumber", zap.Error(err), zap.Uint64("chain", t.ChainId))
					continue
				}
			}
		}
	}
}

type Event struct {
	Table string
	Cols  []dbo.Feild
}

func (ci *ChainIndex) handleNumber(ctx context.Context, cli *ethcli.ETHCli, number uint64, t *model.Task) error {
	block, err := cli.BlockByNumber(ctx, big.NewInt(int64(number)))
	if err != nil {
		return err
	}

	var events []Event

	for _, v := range block.Transactions() {
		if v.To() == nil || v.To().Hex() != common.HexToAddress(t.Contract).String() {
			continue
		}

		if receipt, err := cli.TransactionReceipt(ctx, v.Hash()); err != nil {
			return err
		} else {
			ins, _ := abi.JSON(strings.NewReader(t.Abi))
			for _, v := range receipt.Logs {
				event, err := ins.EventByID(v.Topics[0])
				if err != nil {
					return err
				}

				indexed := v.Topics[1:]
				unindexed, err := event.Inputs.Unpack(v.Data)
				if err != nil {
					return err
				}

				var cols []dbo.Feild
				for i, arg := range event.Inputs {
					if arg.Indexed {
						data := indexed[i].Bytes()
						var value interface{}
						switch arg.Type.T {
						case abi.AddressTy:
							var x common.Address
							x.SetBytes(data)
							value = x.Hex()
						case abi.IntTy, abi.UintTy:
							var x = new(big.Int)
							x.SetBytes(data)
							value = x.String()
						case abi.BoolTy:
							var x bool
							if data[31] == 1 {
								x = true
							}
							value = x
							//todo
						}
						cols = append(cols, dbo.Feild{
							Name:  arg.Name,
							Value: value,
						})
						continue
					}

					// unindexed
					switch reflect.TypeOf(unindexed[i-len(indexed)]).String() {
					case "*big.Int":
						var x *big.Int
						x = unindexed[i-len(indexed)].(*big.Int)
						cols = append(cols, dbo.Feild{
							Name:  arg.Name,
							Value: x.String(),
						})
					default:
						cols = append(cols, dbo.Feild{
							Name:  arg.Name,
							Value: unindexed[i-len(indexed)],
						})
					}
					//cols = append(cols, dbo.Feild{
					//	Name:  arg.Name,
					//	Value: unindexed[i-len(indexed)],
					//})
				}

				events = append(events, Event{
					Table: t.TableName(event.Name),
					Cols:  cols,
				})
			}
		}
	}

	return ci.db.Transaction(func(tx *sql.Tx) error {
		for _, v := range events {
			if _, err := ci.db.Insert(tx, v.Table, v.Cols); err != nil {
				return err
			}
		}
		if _, err := tx.Exec("UPDATE TASK SET CURRENT=CURRENT+1 WHERE ID=?", t.ID); err != nil {
			return err
		}
		t.Current = t.Current + 1
		return nil
	})

}
