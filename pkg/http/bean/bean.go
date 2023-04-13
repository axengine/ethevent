package bean

type PageRo struct {
	Cursor uint64 `query:"cursor" validate:"omitempty"`
	Limit  uint64 `query:"limit" validate:"required,lte=100,gte=1"`
}

type TaskAddRo struct {
	Contract string `query:"contract" validate:"required,len=42,startswith=0x"`
	Abi      string `query:"abi" validate:"required"`
	ChainId  uint64 `query:"chainId" validate:"required,gt=0"`
	Rpc      string `query:"rpc" validate:"required"`
	Start    uint64 `query:"start" validate:"required,gt=0"`
	// 轮询间隔，建议为区块出块间隔
	Interval uint64 `query:"interval" validate:"required,gt=0"`
}

type TaskPauseRo struct {
	Id    uint `json:"id" validate:"required,gt=0"`
	Pause uint `json:"pause" validate:"omitempty"`
}

type TaskDeleteRo struct {
	Id uint `json:"id" validate:"required,gt=0"`
}

type TaskUpdateRo struct {
	Id uint `json:"id" validate:"required,gt=0"`
	TaskAddRo
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
	Op    string `query:"op" validate:"omitempty"` // can be =、>、<、<> and any operator supported by sql-database
}

type OrderRo struct {
	// ASC/DESC
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

type Event map[string]interface{}
