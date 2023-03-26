package chainindex

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/axengine/ethcli"
	"github.com/axengine/ethevent/pkg/database"
	"github.com/axengine/ethevent/pkg/model"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"math/big"
	"strings"
)

type ChainIndex struct {
	db     *database.DBO
	logger *zap.Logger
}

func New(logger *zap.Logger, db *database.DBO) *ChainIndex {
	return &ChainIndex{db: db, logger: logger}
}

// Init initial table and index
func (ci *ChainIndex) Init() error {
	// create table TASK
	if _, err := ci.db.Exec(nil, model.CreateTaskTableSQL); err != nil {
		ci.logger.Error("Exec", zap.Error(err), zap.String("sql", model.CreateTaskTableSQL))
		return err
	}

	var tasks []model.Task
	where := []database.Where{{Name: "1", Value: 1}}
	if err := ci.db.SelectRows("ETH_TASK", where, nil, nil, &tasks); err != nil {
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
				// todo
				switch arg.Type.T {
				case abi.IntTy, abi.UintTy:
					tp = "TEXT"
				case abi.BoolTy:
					tp = "BOOLEAN"
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
				ci.logger.Error("Exec", zap.Error(err), zap.String("sql", ctsqls))
				return err
			}

			cisqls := model.CreateIndexSQL(tableName, indexCols)
			for _, v := range cisqls {
				if _, err := ci.db.Exec(nil, v); err != nil {
					ci.logger.Error("Exec", zap.Error(err), zap.String("sql", v))
					return err
				}
			}
		}
	}

	return nil
}

func (ci *ChainIndex) Start(ctx context.Context) error {
	var tasks []model.Task
	where := []database.Where{{Name: "1", Value: 1}}
	if err := ci.db.SelectRows("ETH_TASK", where, nil, nil, &tasks); err != nil {
		return err
	}

	for _, v := range tasks {
		if cli, err := ethcli.New(v.Rpc); err != nil {
			return err
		} else {
			go ci.start(ctx, cli, &v)
		}
	}
	return nil
}

func (ci *ChainIndex) start(ctx context.Context, cli *ethcli.ETHCli, t *model.Task) {
	for {
		select {
		case <-ctx.Done():
			ci.logger.Info("ChainIndex exit")
			return
		default:
			number, err := cli.BlockNumber(ctx)
			if err != nil {
				ci.logger.Warn("BlockNumber", zap.Error(err))
				continue
			}
			if t.Current < number {
				if err := ci.handleNumber(ctx, cli, t.Current+1, t); err != nil {
					ci.logger.Error("handleNumber", zap.Error(err), zap.Uint64("chain", t.ChainId))
					continue
				}
			}
		}
	}
}

type Event struct {
	Table string
	Cols  []database.Feild
}

func (ci *ChainIndex) handleNumber(ctx context.Context, cli *ethcli.ETHCli, number uint64, t *model.Task) error {
	block, err := cli.BlockByNumber(ctx, big.NewInt(int64(number)))
	if err != nil {
		return err
	}

	var events []Event

	for _, tx := range block.Transactions() {
		if tx.To() == nil || tx.To().Hex() != common.HexToAddress(t.Contract).String() {
			continue
		}

		if receipt, err := cli.TransactionReceipt(ctx, tx.Hash()); err != nil {
			return err
		} else {
			ins, _ := abi.JSON(strings.NewReader(t.Abi))
			for _, rcptLog := range receipt.Logs {
				event, err := ins.EventByID(rcptLog.Topics[0])
				if err != nil {
					continue
				}

				indexed := rcptLog.Topics[1:]
				unindexed, err := event.Inputs.Unpack(rcptLog.Data)
				if err != nil {
					return err
				}

				var cols []database.Feild
				{
					cols = append(cols, database.Feild{
						Name:  "Address",
						Value: rcptLog.Address.Hex(),
					})
					//cols = append(cols, database.Feild{
					//	Name:  "Topics",
					//	Value: v.Topics,
					//})
					//cols = append(cols, database.Feild{
					//	Name:  "Data",
					//	Value: v.Data,
					//})
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
						Value: block.Time(),
					})
					cols = append(cols, database.Feild{
						Name:  "TxHash",
						Value: rcptLog.TxHash.Hex(),
					})
					cols = append(cols, database.Feild{
						Name:  "TxIndex",
						Value: rcptLog.TxIndex,
					})

					//cols = append(cols, database.Feild{
					//	Name:  "Index",
					//	Value: v.Index,
					//})
					cols = append(cols, database.Feild{
						Name:  "Removed",
						Value: rcptLog.Removed,
					})
				}
				var indexedParams = make(map[string]interface{})
				if err := abi.ParseTopicsIntoMap(indexedParams, event.Inputs[:len(indexed)], indexed); err != nil {
					return err
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

				for i, v := range event.Inputs.NonIndexed() {
					if vv, ok := unindexed[i].(fmt.Stringer); ok {
						unindexed[i] = vv.String()
					}
					cols = append(cols, database.Feild{
						Name:  v.Name,
						Value: unindexed[i],
					})
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
		if _, err := tx.Exec("UPDATE ETH_TASK SET CURRENT=CURRENT+1 WHERE ID=?", t.ID); err != nil {
			return err
		}
		t.Current = t.Current + 1
		return nil
	})

}
