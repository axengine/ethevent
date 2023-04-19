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
	cli, _ := ethcli.New("https://goerli.infura.io/v3/03d2548af36149abb66a54983ea238f9")
	block, err := cli.BlockByNumber(context.Background(), big.NewInt(8577205))
	if err != nil {
		t.Fatal(err)
	}

	evABI, _ := abi.JSON(strings.NewReader(`[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`))

	for k, v := range evABI.Events {
		fmt.Println("k=", k, " v=", v.ID.Hex(), " matched=", block.Bloom().Test(v.ID.Bytes()))
	}

	{
		b := block.Bloom().Test(common.HexToAddress("0xC01138c43c8D99732fa900059FCAA9f34Cd6047a").Bytes())
		fmt.Println(b)
	}

	for _, v := range block.Transactions() {
		receipt, err := cli.TransactionReceipt(context.Background(), v.Hash())
		if err != nil {
			t.Fatal(err)
		}
		testInReceipt := receipt.Bloom.Test(common.HexToAddress("0xC01138c43c8D99732fa900059FCAA9f34Cd6047a").Bytes())
		fmt.Println(testInReceipt, " ", v.Hash().Hex())
	}
}
