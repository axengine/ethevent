package svc

import (
	"context"
	"fmt"
	"github.com/axengine/ethevent/pkg/database"
	"github.com/axengine/ethevent/pkg/http/bean"
	"github.com/axengine/ethevent/pkg/model"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strings"
	"time"
)

type Service struct {
	db     *database.DBO
	logger *zap.Logger
}

func New(logger *zap.Logger, db *database.DBO) *Service {
	return &Service{logger: logger, db: db}
}

func (svc *Service) TaskList(cursor, limit uint64, order string) ([]model.Task, error) {
	var datas []model.Task
	where := []database.Where{
		database.Where{Name: "1", Value: 1},
	}
	o, err := database.MakeOrder(order, "Id")
	if err != nil {
		return nil, err
	}
	p := database.MakePaging("id", cursor, limit)
	if err := svc.db.SelectRows("ETH_TASK", where, o, p, &datas); err != nil {
		return nil, err
	}
	return datas, nil
}

func (svc *Service) TaskAdd(ctx context.Context, req *bean.TaskAddRo) error {
	if _, err := abi.JSON(strings.NewReader(req.Abi)); err != nil {
		return err
	}

	fields := []database.Feild{
		database.Feild{Name: "contract", Value: common.HexToAddress(req.Contract).Hex()},
		database.Feild{Name: "abi", Value: req.Abi},
		database.Feild{Name: "chainId", Value: req.ChainId},
		database.Feild{Name: "rpc", Value: req.Rpc},
		database.Feild{Name: "start", Value: req.Start},
		database.Feild{Name: "current", Value: req.Start},
		database.Feild{Name: "interval", Value: req.Interval},
	}

	_, err := svc.db.Insert(nil, "ETH_TASK", fields)
	return err
}

func (svc *Service) TaskUpdate(ctx context.Context, req *bean.TaskUpdateRo) error {
	if _, err := abi.JSON(strings.NewReader(req.Abi)); err != nil {
		return err
	}

	where := []database.Where{
		database.Where{Name: "id", Value: req.Id},
	}

	fields := []database.Feild{
		database.Feild{Name: "contract", Value: common.HexToAddress(req.Contract).Hex()},
		database.Feild{Name: "abi", Value: req.Abi},
		database.Feild{Name: "chainId", Value: req.ChainId},
		database.Feild{Name: "rpc", Value: req.Rpc},
		database.Feild{Name: "start", Value: req.Start},
		database.Feild{Name: "interval", Value: req.Interval}}

	_, err := svc.db.Update(nil, "ETH_TASK", fields, where)
	return err
}

func (svc *Service) TaskPause(ctx context.Context, req *bean.TaskPauseRo) error {
	where := []database.Where{
		database.Where{Name: "id", Value: req.Id},
	}
	_, err := svc.db.Update(nil, "ETH_TASK", []database.Feild{
		database.Feild{
			Name:  "paused",
			Value: req.Pause,
		},
	}, where)
	return err
}

func (svc *Service) TaskDelete(ctx context.Context, req *bean.TaskDeleteRo) error {
	where := []database.Where{
		database.Where{Name: "id", Value: req.Id},
	}
	_, err := svc.db.Update(nil, "ETH_TASK", []database.Feild{
		database.Feild{
			Name:  "deletedAt",
			Value: time.Now().Unix(),
		},
	}, where)
	return err
}

func (svc *Service) EventList(req *bean.EventListRo) ([]map[string]interface{}, error) {
	task, err := svc.findTaskByContract(req.Contract)
	if err != nil {
		return nil, err
	}
	if task == nil || task.ID == 0 {
		return nil, errors.New("Not found task")
	}

	tableName := fmt.Sprintf("EVENT_%d_%s", task.ID, req.Event)

	where := []database.Where{
		database.Where{Name: "1", Value: 1},
	}
	for _, v := range req.Where {
		if v.Name != "" && v.Value != "" {
			where = append(where, database.Where{Name: v.Name, Value: v.Value})
		}
	}

	if req.BlockRo != nil {
		if req.BlockRo.Number > 0 {
			where = append(where, database.Where{Name: "BlockNumber", Value: req.BlockRo.Number})
		}
		if req.BlockRo.Hash != "" {
			where = append(where, database.Where{Name: "BlockHash", Value: req.BlockRo.Hash})
		}
	}
	if req.TxRo != nil {
		if req.TxRo.Hash != "" {
			where = append(where, database.Where{Name: "TxHash", Value: req.TxRo.Hash})
		}
	}
	if req.TimeRo != nil {
		if req.TimeRo.Begin > 0 {
			where = append(where, database.Where{Name: "BlockTime", Value: req.TimeRo.Begin, Op: ">="})
		}
		if req.TimeRo.End > 0 {
			where = append(where, database.Where{Name: "BlockTime", Value: req.TimeRo.End, Op: "<"})
		}
	}

	orderT, _ := database.MakeOrder("", "ID")

	var paging *database.Paging
	if req.PageRo != nil {
		paging = database.MakePaging("ID", req.PageRo.Cursor, req.PageRo.Limit)
	}

	return svc.db.SelectRowsToMaps(tableName, where, orderT, paging)
}

func (svc *Service) findTaskByContract(contract string) (*model.Task, error) {
	var datas []model.Task
	where := []database.Where{
		database.Where{Name: "Contract", Value: contract},
	}

	if err := svc.db.SelectRows("ETH_TASK", where, nil, nil, &datas); err != nil {
		return nil, err
	}
	if len(datas) > 0 {
		return &datas[0], nil
	}
	return nil, nil
}
