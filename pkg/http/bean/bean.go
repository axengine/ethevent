package bean

type PageRo struct {
	Cursor uint64 `query:"cursor" validate:"required,gte=0"`
	Limit  uint64 `query:"limit" validate:"required,lte=100,gte=1"`
}

type TaskAddRo struct {
	Contract string `query:"contract" validate:"required,len=42,startswith=0x"`
	Abi      string `query:"abi" validate:"required"`
	ChainId  uint64 `query:"chainId" validate:"required,gt=0"`
	Rpc      string `query:"rpc" validate:"required"`
	Begin    uint64 `query:"begin" validate:"required,gt=0"`
}

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

type EventListRo struct {
	Contract string   `query:"contract" json:"contract" validate:"required,len=42,startswith=0x"`
	Event    string   `query:"event" json:"event" validate:"required,gt=0"`
	Where    []Where  `query:"where" json:"where" validate:"omitempty"`
	BlockRo  *BlockRo `query:"blockRo" json:"blockRo" validate:"omitempty"`
	TxRo     *TxRo    `query:"txRo" json:"txRo" validate:"omitempty"`
	TimeRo   *TimeRo  `query:"timeRo" json:"timeRo" validate:"omitempty"`
	PageRo   *PageRo  `query:"pageRo" json:"pageRo" validate:"required"`
}

//type EventListQueryRo struct {
//	Contract    string `query:"contract" json:"contract" validate:"required,len=42,startswith=0x"`
//	Event       string `query:"event" json:"event" validate:"required"`
//	BlockNumber uint64 `query:"blockNumber" validate:"omitempty"`
//	BlockHash   string `query:"blockHash" validate:"omitempty"`
//	TxHash      string `query:"txHash" validate:"omitempty"`
//	Begin       int64  `query:"begin" validate:"omitempty"`
//	End         int64  `query:"end" validate:"omitempty"`
//	Cursor      uint64 `query:"cursor" validate:"required,gte=0"`
//	Limit       uint64 `query:"limit" validate:"required,lte=100,gte=1"`
//
//	// name && value 作为条件 可以传数组
//	Name  string `query:"name" validate:"omitempty"`
//	Value string `query:"value" validate:"omitempty"`
//}

type Event map[string]interface{}
