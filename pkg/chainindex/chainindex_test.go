package chainindex

import (
	"context"
	"fmt"
	"github.com/axengine/ethcli"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func TestType(t *testing.T) {
	var value = big.NewInt(123)
	fn := func(v interface{}) {
		typeOfV := reflect.TypeOf(value)
		fmt.Println(typeOfV.String())
		switch typeOfV.String() {
		case "*big.Int":
			fmt.Println("*big.Int")
		}
	}

	fn(value)
}

func TestBloom(t *testing.T) {
	cli, _ := ethcli.New("https://bsc-dataseed1.ninicoin.io/")
	block, err := cli.BlockByNumber(context.Background(), big.NewInt(27471643))
	if err != nil {
		t.Fatal(err)
	}

	evABI, _ := abi.JSON(strings.NewReader(`[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`))

	for k, v := range evABI.Events {
		fmt.Println("k=", k, " v=", v.ID.Hex(), " matched=", block.Bloom().Test(v.ID.Bytes()))
	}

	{
		b := block.Bloom().Test(common.HexToAddress("0xc748673057861a797275cd8a068abb95a902e8de").Bytes())
		fmt.Println(b)
	}
	{
		b := block.Bloom().Test(common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef").Bytes())
		fmt.Println(b)
	}
	{
		b := block.Bloom().Test(common.HexToHash("0x643337e906111d248a5907ab20e25dac96b6ea69").Bytes())
		fmt.Println(b)
	}
	{
		b := block.Bloom().Test(common.HexToHash("0xc736ca3d9b1e90af4230bd8f9626528b3d4e0ee0").Bytes())
		fmt.Println(b)
	}
	for _, v := range block.Transactions() {
		receipt, err := cli.TransactionReceipt(context.Background(), v.Hash())
		if err != nil {
			t.Fatal(err)
		}
		testInReceipt := receipt.Bloom.Test(common.HexToAddress("0xc748673057861a797275cd8a068abb95a902e8de").Bytes())
		fmt.Println(testInReceipt, " ", v.Hash().Hex())
	}
}
