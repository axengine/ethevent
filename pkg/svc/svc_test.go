package svc

import (
	"encoding/json"
	"fmt"
	"github.com/axengine/ethevent/pkg/database"
	"github.com/axengine/ethevent/pkg/http/bean"
	"github.com/axengine/utils/log"
	"os"
	"path/filepath"
	"testing"
)

var tSVC *Service

func TestMain(m *testing.M) {
	dbo := database.New(filepath.Join("../../", "events.db"), log.Logger)
	tSVC = New(log.Logger, dbo)
	os.Exit(m.Run())
}

func TestValueDecimal(t *testing.T) {
	var evReq = &bean.EventListRo{
		TaskId: 1,
		Event:  "Transfer",
		Cols: []string{"ID",
			"Address",
			"BlockNumber",
			"BlockHash",
			"BlockTime",
			"TxHash",
			"TxIndex",
			"Method",
			"[FROM]",
			"[TO]",
			"CAST(VALUE/1e18 as decimal(20,8)) AS VALUE1",
			"ROUND(VALUE/1e18,6) AS VALUE",
		},
		OrderRo: &bean.OrderRo{
			OrderType: "DESC",
			Feilds:    []string{"ID"},
		},
		PageRo: &bean.PageRo{
			Cursor: 0,
			Limit:  100,
		},
	}
	evReq.Wheres = append(evReq.Wheres, []bean.Where{{
		Name: "TxHash",
		// 0x200aa47d562f27149affb9c9b8d12303c2675f43f92e9f57f7f1b65619ffaa54
		Value: "0x18307a4fa2a6966dc8fbabea36fbac5dddc6cd1f7b6c8cb8bb1a6e2a170b4a57",
		Op:    "=",
	}})
	out, err := tSVC.EventList(evReq)
	if err != nil {
		t.Fatal(err)
	}
	bz, _ := json.MarshalIndent(out, "", " ")
	fmt.Println(string(bz))
}
