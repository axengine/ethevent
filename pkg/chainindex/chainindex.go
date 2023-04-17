package chainindex

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"github.com/axengine/ethcli"
	"github.com/axengine/ethevent/pkg/database"
	"github.com/axengine/ethevent/pkg/model"
	"github.com/axengine/utils/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math/big"
	"strings"
	"sync"
	"time"
)

type ChainIndex struct {
	db     *database.DBO
	logger *zap.Logger

	mu    sync.RWMutex
	tasks map[uint]*model.Task
}

func New(logger *zap.Logger, db *database.DBO) *ChainIndex {
	return &ChainIndex{db: db, logger: logger, tasks: map[uint]*model.Task{}}
}

// Init initial table and index
func (ci *ChainIndex) Init() error {
	// create table TASK
	if _, err := ci.db.Exec(nil, model.CreateTaskTableSQL); err != nil {
		ci.logger.Error("Exec", zap.Error(err), zap.String("sql", model.CreateTaskTableSQL))
		return err
	}

	return nil
}

func (ci *ChainIndex) initTask(task *model.Task) error {
	var tablePrefix = fmt.Sprintf("EVENT_%d_", task.ID)
	ins, err := abi.JSON(strings.NewReader(task.Abi))
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

	return nil
}

func (ci *ChainIndex) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	tm := time.NewTimer(time.Second * 1)
	for {
		select {
		case <-ctx.Done():
			ci.logger.Info("Start exit")
			return
		case <-tm.C:
			var tasks []model.Task
			where := []database.Where{{Name: "1", Value: 1}}
			if err := ci.db.SelectRows("ETH_TASK", nil, where, nil, nil, &tasks); err != nil {
				log.Logger.Error("SelectRows", zap.Error(err))
				return
			}

			for _, v := range tasks {
				task := v
				ci.mu.RLock()
				_, ok := ci.tasks[task.ID]
				ci.mu.RUnlock()
				if ok {
					continue
				}

				if cli, err := ethcli.New(task.Rpc); err != nil {
					log.Logger.Error("ethcli.New", zap.Error(err), zap.String("rpc", task.Rpc))
					continue
				} else {
					ci.mu.Lock()
					ci.tasks[task.ID] = &task
					ci.mu.Unlock()

					if err := ci.initTask(&task); err != nil {
						log.Logger.Error("initTask", zap.Error(err), zap.Uint("task", task.ID))
						continue
					}
					wg.Add(1)
					go ci.start(ctx, wg, cli, &task)
				}
			}
			tm.Reset(time.Second * 5)
		}
	}
}

func (ci *ChainIndex) start(ctx context.Context, wg *sync.WaitGroup, cli *ethcli.ETHCli, t *model.Task) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			ci.logger.Info("ChainIndex exit")
			return
		default:
			var tasks []model.Task
			where := []database.Where{
				{Name: "1", Value: 1},
				{Name: "ID", Value: t.ID},
			}
			if err := ci.db.SelectRows("ETH_TASK", nil, where, nil, nil, &tasks); err != nil {
				ci.logger.Error("SelectRows", zap.Error(err))
				return
			}
			if len(tasks) > 0 {
				t = &tasks[0]
			}
			if t.DeletedAt > 0 {
				ci.logger.Debug("Task Deleted", zap.Uint("id", t.ID))
				return
			}
			if t.Paused == 1 {
				ci.logger.Debug("Task Paused", zap.Uint("id", t.ID))
				time.Sleep(time.Second * 5)
				continue
			}

			number, err := cli.BlockNumber(ctx)
			if err != nil {
				ci.logger.Warn("BlockNumber", zap.Error(err))
				continue
			}
			if t.Current < t.Start {
				t.Current = t.Start
			}
			if t.Current >= number {
				next := time.Unix(t.UpdatedAt+t.Interval, 0)
				time.Sleep(next.Sub(time.Now()))
				continue
			}
			if t.Current < number {
				if err := ci.handleNumber(ctx, cli, t.Current+1, t); err != nil {
					ci.logger.Error("handleNumber", zap.Error(err), zap.Uint64("chain", t.ChainId))
					continue
				}
				ci.logger.Info("handleNumber", zap.Uint("task", t.ID), zap.Uint64("height", t.Current+1))
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
		//if tx.To() == nil || tx.To().Hex() != common.HexToAddress(t.Contract).String() {
		//	continue
		//}

		if receipt, err := cli.TransactionReceipt(ctx, tx.Hash()); err != nil {
			return err
		} else {
			ins, _ := abi.JSON(strings.NewReader(t.Abi))
			for _, rcptLog := range receipt.Logs {
				event, err := ins.EventByID(rcptLog.Topics[0])
				if err != nil {
					continue
				}

				eventAddress := rcptLog.Address.Hex()
				if eventAddress != common.HexToAddress(t.Contract).Hex() {
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
				}

				// 非索引参数
				if len(rcptLog.Data) > 0 {
					unindexed, err := event.Inputs.Unpack(rcptLog.Data)
					if err != nil {
						return errors.Wrap(err, "event.Inputs.Unpack tx:"+tx.Hash().Hex())
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
		if _, err := tx.Exec("UPDATE ETH_TASK SET CURRENT=? ,UpdatedAt=? WHERE ID=?",
			number, time.Now().Unix(), t.ID); err != nil {
			return err
		}
		t.Current = number
		return nil
	})

}
