package chainindex

import (
	"context"
	"fmt"
	"github.com/axengine/ethcli"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"reflect"
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
	cli, _ := ethcli.New("https://endpoints.omniatech.io/v1/bsc/testnet/public")
	block, err := cli.BlockByNumber(context.Background(), big.NewInt(29015631))
	if err != nil {
		t.Fatal(err)
	}
	{
		b := block.Bloom().Test(common.HexToAddress("0xD085CE10bC2055fe8caA0e1137ebb10854E51CB7").Bytes())
		fmt.Println(b)
	}
	{
		b := block.Bloom().Test(common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef").Bytes())
		fmt.Println(b)
	}
	{
		b := block.Bloom().Test(common.HexToHash("0x0000000000000000000000001100e4b8674aea98a2ac239432f41f3bfb50c671").Bytes())
		fmt.Println(b)
	}
	{
		b := block.Bloom().Test(common.HexToHash("0x0000000000000000000000009ee099fb3c185c99b4bdc85c2bed9f5d0b1ced18").Bytes())
		fmt.Println(b)
	}
	//for _, v := range block.Transactions() {
	//	receipt, err := cli.TransactionReceipt(context.Background(), v.Hash())
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	testInReceipt := receipt.Bloom.Test(common.HexToAddress("0xD085CE10bC2055fe8caA0e1137ebb10854E51CB7").Bytes())
	//	fmt.Println(testInReceipt, " ", v.Hash().Hex())
	//}
}
