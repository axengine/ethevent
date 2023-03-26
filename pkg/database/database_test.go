package database

import (
	"fmt"
	"github.com/axengine/ethevent/pkg/http/bean"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestName(t *testing.T) {
	conn, err := sqlx.Connect("sqlite3", "./test1.db")
	if err != nil {
		t.Fatal(err)
	}
	//var data = make(map[string]interface{})
	var data []bean.Event
	rows, err := conn.Queryx("select * from EVENT_1_Transfer where 1 = 1 order by ID desc")
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var dest = make(map[string]interface{})
		if err = rows.MapScan(dest); err != nil {
			t.Fatal(err)
		}
		fmt.Println(dest)
		data = append(data, dest)
	}
	fmt.Println(data)
}
