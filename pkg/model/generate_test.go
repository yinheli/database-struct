package model

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerate(t *testing.T) {
	option := &Options{
		DbType:          DbTypeMySQL,
		Dsn:             "root:CHEN2013vmovierH@(192.168.4.200:3306)/mod_v2?charset=utf8mb4&parseTime=True&loc=Local",
		GenGormTag:      true,
		GenJsonTag:      true,
		HtmlFile:        "/Users/yinheli/Downloads/tmp/tmp.html",
		ModelDir:        "/Users/yinheli/Downloads/tmp",
		ModelSingleFile: true,
	}
	tables, err := DbStruct(option)
	require.NoError(t, err)
	err = Generate(option, tables)
	require.NoError(t, err)
}

func TestDbStruct(t *testing.T) {
	option := &Options{
		DbType:     DbTypeMySQL,
		Dsn:        "root:CHEN2013vmovierH@(192.168.4.200:3306)/mod_v2?charset=utf8mb4&parseTime=True&loc=Local",
		GenGormTag: true,
		GenJsonTag: true,
	}

	tables, err := DbStruct(option)
	require.NoError(t, err)
	t.Log("table count:", len(tables))
}

func Test_jen(t *testing.T) {
	c := jen.Comment("aa").Line().Type().Id("hello").Struct(
		jen.Id("Name").Op("*").String().Comment("hello"),
		jen.Id("Byte").Op("[]").Byte().Comment("hello"),
		jen.Id("t").Qual("time", "Time"),
	)

	fmt.Println(c.GoString())
}
