# ethevent
Parse and store events in the Ethereum-like blockchain, and query them through REST

ethevent是一个简版的graph-node，用户添加任务（要解析的合约事件），ethevent解析合约事件存储到sqlite3，提供http查询接口查询事件列表。

## usage
`./build/ethevent --datadir=. --http.port=8080`

via:http://localhost:8080/docs/index.html

## 任务管理
- /v1/task/add 添加一个任务
- /v1/task/pause 任务暂停或恢复
- /v1/task/delete 删除任务
- /v1/task/update 更新任务
- /v1/task/list 查询任务列表

## 事件查询
- /v1/event/list 查询事件日志

## example
- Add a task:在BSC上解析USDT的Transfer事件，从26800040开始解析，解析间隔3秒
```shell
curl 'http://localhost:8080/v1/task/add' \
	-H "Content-Type:application/json" \
	-X POST \
	-d '{"abi":"[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"}]","chainId":56,"contract":"0x55d398326f99059fF775485246999027B3197955","interval":3,"rpc":"https://bsc-dataseed1.ninicoin.io/","start":26800040}'

{"code":0,"msg":"ok","data":1}
```

- List event logs:查询任务Id为`1`的`Transfer`事件，分页采用游标，默认倒序，更多参数查看swag文档
```shell
curl 'http://localhost:8080/v1/event/list' \
	-H "Content-Type:application/json" \
	-X POST \
	-d '{"taskId":1,"event":"Transfer","pageRo":{"cursor":0,"limit":100}}'
```

## /v1/event/list request
```
type BlockRo struct {
	Number uint64 `query:"number" json:"number" validate:"omitempty"`
	Hash   string `query:"hash" json:"hash" validate:"omitempty"`
}

type TxRo struct {
	Hash string `query:"hash" json:"hash" validate:"omitempty"`
}

type TimeRo struct {
	Begin int64 `query:"begin" json:"begin" validate:"omitempty"`
	End   int64 `query:"end" json:"end" validate:"omitempty"`
}

type Where struct {
	Name  string      `query:"name" json:"name" validate:"omitempty"`
	Value interface{} `query:"value" json:"value" validate:"omitempty"`
	Op    string      `query:"op" json:"op" validate:"omitempty"` // can be =、>、<、<> and any operator supported by sql-database
}

type OrderRo struct {
	OrderType string   `query:"orderType" json:"orderType" validate:"omitempty,oneof=ASC DESC"`
	Feilds    []string `query:"feilds" json:"feilds" validate:"omitempty"`
}

type EventListRo struct {
	TaskId  uint      `query:"taskId" json:"taskId" validate:"required,gt=0"`
	Event   string    `query:"event" json:"event" validate:"required,gt=0"`
	Cols    []string  `query:"cols" json:"cols" validate:"omitempty"`
	Wheres  [][]Where `query:"wheres" json:"wheres" validate:"omitempty"`
	BlockRo *BlockRo  `query:"blockRo" json:"blockRo" validate:"omitempty"`
	TxRo    *TxRo     `query:"txRo" json:"txRo" validate:"omitempty"`
	TimeRo  *TimeRo   `query:"timeRo" json:"timeRo" validate:"omitempty"`
	OrderRo *OrderRo  `query:"orderRo" json:"orderRo" validate:"omitempty"`
	PageRo  *PageRo   `query:"pageRo" json:"pageRo" validate:"required"`
}
```
- taskId 和 event是要求必传的，通过taskId和event确定查询的表
- cols可以指定要查询的字段，所有事件除了事件本身的字段外，还包含以下公共字段:
- cols公共字段：`ID` `Address` `BlockHash` `BlockNumber` `BlockTime` `TxHash` `TxIndex` `Method`
- cols支持SQL可查询字段，例如：`sum(value)`,`avg(value)`,`count(*)`等sql可查询的字段
- blockRo 区块高度和hash 作为公共条件
- timeRo 区块时间 作为公共条件
- txRo 交易hash 作为公共条件
- orderRo 排序，指定排序字段和方式
- pageRo 当出现orderRo时pageRo必传，游标分页；单次查询最大100条数据,防止查询大量数据带来的负荷
- wheres 指定自定义查询条件，如果len(wheres)>1，将执行union查询,此时 blockRo/timeRo/txRo 将作为公共条件；
```
{{A}, {B}}         matches where A union where B
{{A, B}, {C, D}} matches where (A AND B) union where (C AND D)
{{A, B}, {C, D}, {E, F}} matches where (A AND B) union where (C AND D) union where (E AND F)
```
- cols和wheres中指定字段名称时，如果是sql关键字，需要用[]括起来，例如:`[FROM]` `[TO]`

### request example
- 查询记录
```json
{
  "taskId": 1,
  "event": "Transfer",
  "cols": [
    "ID",
    "Address",
    "BlockNumber",
    "BlockHash",
    "BlockTime",
    "TxHash",
    "TxIndex",
    "Method",
    "[FROM]",
    "[TO]",
    "VALUE"
  ],
  "wheres": [
    [
      {
        "name": "[FROM]",
        "value": "0x1100e4B8674aea98a2AC239432f41f3BFB50c671",
        "op": "="
      }
    ],
    [
      {
        "name": "[TO]",
        "value": "0x3d2e7a5ffFa8eBc7C82C4327B605c4a7DDb714Db",
        "op": "="
      }
    ]
  ],
  "blockRo": {
    "number": 0,
    "hash": ""
  },
  "txRo": {
    "hash": ""
  },
  "timeRo": {
    "begin": 1681736896,
    "end": 1681736899
  },
  "pageRo": {
    "cursor": 0,
    "limit": 100
  },
  "orderRo": {
    "orderType": "DESC",
    "feilds": [
      "ID",
      "BlockNumber"
    ]
  }
}
```
```
 {"sql": "select ID,Address,BlockNumber,BlockHash,BlockTime,TxHash,TxIndex,Method,[FROM],[TO],VALUE from EVENT_1_Transfer where 1 = 1 and [FROM] = ?  and BlockTime >= ?  and BlockTime < ?  
 union select ID,Address,BlockNumber,BlockHash,BlockTime,TxHash,TxIndex,Method,[FROM],[TO],VALUE from EVENT_1_Transfer where 1 = 1 and [TO] = ?  and BlockTime >= ?  and BlockTime < ?  
 order by ID  , BlockNumber DESC limit 100 offset 0 ", 
 "values": ["0x1100e4B8674aea98a2AC239432f41f3BFB50c671",1681736896,1681736899,"0x3d2e7a5ffFa8eBc7C82C4327B605c4a7DDb714Db",1681736896,1681736899]}
```

- 统计
```json
{
  "taskId": 1,
  "event": "Transfer",
  "cols":["AVG(VALUE)"]
}
```
```
{"sql": "select AVG(VALUE) from EVENT_1_Transfer where 1 = 1 and 1 = ? ", "values": [1]}
```