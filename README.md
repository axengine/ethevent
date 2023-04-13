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

{"resCode":0,"resDesc":"ok","result":1}
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
	Number uint64 `query:"number" validate:"omitempty"`
	Hash   string `query:"hash" validate:"omitempty"`
}

type TxRo struct {
	Hash string `query:"hash" validate:"omitempty"`
}

type TimeRo struct {
	Begin int64 `query:"begin" validate:"omitempty"`
	End   int64 `query:"end" validate:"omitempty"`
}

type Where struct {
	Name  string `query:"name" validate:"omitempty"`
	Value string `query:"value" validate:"omitempty"`
}

type OrderRo struct {
	OrderType string   `query:"orderType" validate:"omitempty,oneof=ASC DESC"`
	Feilds    []string `query:"feilds" validate:"omitempty"`
}

type EventListRo struct {
	TaskId  uint     `query:"taskId" json:"taskId" validate:"required,gt=0"`
	Event   string   `query:"event" json:"event" validate:"required,gt=0"`
	Cols    []string `query:"cols" json:"cols" validate:"omitempty"`
	Where   []Where  `query:"where" json:"where" validate:"omitempty"`
	BlockRo *BlockRo `query:"blockRo" json:"blockRo" validate:"omitempty"`
	TxRo    *TxRo    `query:"txRo" json:"txRo" validate:"omitempty"`
	TimeRo  *TimeRo  `query:"timeRo" json:"timeRo" validate:"omitempty"`
	PageRo  *PageRo  `query:"pageRo" json:"pageRo" validate:"required"`
	OrderRo *OrderRo `query:"orderRo" json:"orderRo" validate:"omitempty"`
}
```
- taskId 和 event是要求必传的，pageRo只在查询`记录`时用于分页，是必传项。
- cols可以指定要查询的字段，所有事件初了事件本身的字段外，还包含以下公共字段:
```
type EventBase struct {
	ID      int64  `json:"id"`
	Address string `json:"address"`
	//Topics      []string `json:"topics"`
	//Data        []byte   `json:"data"`
	BlockNumber uint64 `json:"blockNumber"`
	BlockHash   string `json:"blockHash"`
	BlockTime   int64  `json:"blockTime"`
	TxHash      string `json:"txHash"`
	TxIndex     uint   `json:"txIndex"`
	//LogIndex uint `json:"logIndex"`
	Removed bool `json:"removed"`
}
```
cols字段支持统计字段，例如：`sum(value)`,`avg(value)`,`count(*)`等sql可查询的字段
- where 指定自定义查询条件，例如:Name="from" Value="0xe5DaF2824B43d8b0C961225Ab9992baf39F5F835" Op="="
- blockRo 区块高度和hash
- timeRo 区块时间
- txRo 交易hash
- pageRo 游标分页
- orderRo 排序，默认`ID DESC`