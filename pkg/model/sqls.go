package model

import (
	"fmt"
	"strings"
)

var CreateTaskTableSQL = `CREATE TABLE IF NOT EXISTS TASK (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    contract  TEXT    NOT NULL,
    abi       TEXT    NOT NULL,
    chainId   INTEGER DEFAULT (0),
    rpc       TEXT    NOT NULL,
    [begin]   INTEGER DEFAULT (1),
    [current] INTEGER DEFAULT (1) 
);
`

func CreateTableSQL(tableTame string, cols []string) string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
	(
		ID          INTEGER      PRIMARY KEY AUTOINCREMENT,
		Address     STRING (42),
		Topics      TEXT,
		Data        BLOB,
		BlockHash   VARCHAR (66),
		BlockNumber INTEGER,
		BlockTime   TIME,
		TxHash      VARCHAR (66),
		TxIndex     INTEGER,
		LogIndex    INTEGER,
		Removed     BOOLEAN,
    %s
	);`, tableTame, strings.Join(cols, ","))
}

func CreateIndexSQL(tableName string, cols []string) []string {
	var indexSQLs []string
	indexSQLs = append(indexSQLs, fmt.Sprintf(`CREATE INDEX IF NOT EXISTS IDX_%s_%s ON %s (%s);`, tableName, "BlockTime", tableName, "BlockTime"))
	indexSQLs = append(indexSQLs, fmt.Sprintf(`CREATE INDEX IF NOT EXISTS IDX_%s_%s ON %s (%s);`, tableName, "BlockNumber", tableName, "BlockNumber"))
	indexSQLs = append(indexSQLs, fmt.Sprintf(`CREATE INDEX IF NOT EXISTS IDX_%s_%s ON %s (%s);`, tableName, "TxHash", tableName, "TxHash"))

	for _, v := range cols {
		indexSQLs = append(indexSQLs, fmt.Sprintf(`CREATE INDEX IF NOT EXISTS IDX_%s_%s ON %s ("%s");`, tableName, v, tableName, v))
	}
	return indexSQLs
}
