package bean

import (
	"fmt"
	"testing"
)

func TestLang(t *testing.T) {
	accept := "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6"
	tags := parseTags("", accept)
	for _, v := range tags {
		fmt.Println(v.String())
	}
}
