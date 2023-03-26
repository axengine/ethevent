package database

import "errors"

// Feild database field
type Feild struct {
	Name  string
	Value interface{}
}

// Where query field
type Where struct {
	Name  string
	Value interface{}
	Op    string // can be =、>、<、<> and any operator supported by sql-database
}

// GetOp get operator of current where clause, default =
func (w *Where) GetOp() string {
	if w.Op == "" {
		return "="
	}
	return w.Op
}

// Order  used to identify query order
type Order struct {
	Type   string   // "asc" or "desc"
	Feilds []string // order by x
}

// GetOp used in sql
func (o *Order) GetOp() string {
	if o != nil && o.Type == "desc" {
		return "<="
	}

	return ">="
}

type Paging struct {
	CursorName  string // cursor column
	CursorValue uint64 // cursor column
	Limit       uint64 // limit
}

// MakeOrder make a order object
func MakeOrder(ordertype string, fields ...string) (*Order, error) {
	if ordertype == "" {
		ordertype = "desc"
	}

	if ordertype != "asc" && ordertype != "ASC" && ordertype != "desc" && ordertype != "DESC" {
		return nil, errors.New("invalid order type :" + ordertype)
	}

	return &Order{
		Type:   ordertype,
		Feilds: fields,
	}, nil
}

// MakePaging make a paging object
func MakePaging(colName string, colValue uint64, limit uint64) *Paging {
	if limit == 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}
	if colValue < 0 {
		colValue = 0
	}

	return &Paging{
		CursorName:  colName,
		CursorValue: colValue * limit,
		Limit:       limit,
	}
}
