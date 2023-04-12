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
