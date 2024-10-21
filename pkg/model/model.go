package model

import "github.com/dave/jennifer/jen"

type Table struct {
	Ddl         string
	Prefix      string
	Name        string
	Comment     string
	Fields      []*Field
	GoStruct    string
	goStatement *jen.Statement
}

type Field struct {
	Field    string
	Type     string
	Null     string
	Key      string
	Default  *string
	Comment  string
	Nullable bool
	GoType   string
}
