package model

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"regexp"
	"strings"
)

type mysql struct{}

func (t *mysql) dbStruct(options *Options) (tables []*Table, err error) {
	var db *gorm.DB
	db, err = newDb(options.DbType, options.Dsn)
	if err != nil {
		return
	}

	if len(options.Filters) > 0 {
		tables = make([]*Table, 0, 1024)
		nameSet := make(map[string]bool)

		for _, filter := range options.Filters {
			var tbs []*Table
			tbs, err = t.filterTables(db, filter, options.Exclude)
			if err != nil {
				return
			}

			for _, table := range tbs {
				if _, ok := nameSet[table.Name]; ok {
					continue
				}
				nameSet[table.Name] = true
				tables = append(tables, table)
			}
		}
	} else {
		tables, err = t.filterTables(db, nil, options.Exclude)
	}

	return
}

func (t *mysql) filterTables(db *gorm.DB, filter *Filter, exclude []string) (tables []*Table, err error) {
	type mysqlTable struct {
		Name    string `gorm:"column:table_name"`
		Comment string `gorm:"column:table_comment"`
	}

	var dbTables []*mysqlTable

	tdb := db.Table("information_schema.tables").
		Select("table_name, table_comment").
		Where("table_schema = database()")

	if filter != nil {
		tdb = tdb.Where("table_name like ?", filter.TableNamePattern)
	}

	if len(exclude) > 0 {
		tdb.Where("table_name not in(?)", exclude)
	}

	err = tdb.Find(&dbTables).Error
	if err != nil {
		return
	}

	tables = make([]*Table, 0, len(dbTables))

	for _, it := range dbTables {
		tb := &Table{
			Name:    it.Name,
			Comment: it.Comment,
		}

		tb.Ddl = t.tableDdl(db, it.Name)
		tb.Fields, err = t.tableFields(db, it.Name)
		if err != nil {
			return
		}

		if filter != nil {
			tb.Prefix = filter.TablePrefix
		}

		tables = append(tables, tb)
	}

	return
}

func (t *mysql) tableDdl(db *gorm.DB, name string) (ddl string) {
	row := db.Raw(fmt.Sprint("show create table ", name)).Row()
	if db.Error != nil {
		return
	}
	_ = row.Scan(new(string), &ddl)
	return
}

func (t *mysql) tableFields(db *gorm.DB, name string) (fields []*Field, err error) {
	type mysqlField struct {
		ColumnName    string `gorm:"column:column_name"`
		ColumnDefault string `gorm:"column:column_default"`
		IsNullable    string `gorm:"column:is_nullable"`
		DataType      string `gorm:"column:data_type"`
		ColumnType    string `gorm:"column:column_type"`
		ColumnKey     string `gorm:"column:column_key"`
		Extra         string `gorm:"column:extra"`
		ColumnComment string `gorm:"column:column_comment"`
	}

	var dbFields []*mysqlField

	fdb := db.Table("information_schema.columns").
		Select("column_name, column_default, is_nullable, data_type, column_type, column_key, extra, column_comment").
		Where("table_schema=database() and table_name=?", name)
	err = fdb.Find(&dbFields).Error
	if err != nil {
		return
	}

	fields = make([]*Field, 0, len(dbFields))
	for _, it := range dbFields {
		field := &Field{
			Field:   it.ColumnName,
			Type:    strings.ToLower(it.ColumnType),
			Null:    strings.ToUpper(it.IsNullable),
			Key:     it.ColumnKey,
			Default: it.ColumnDefault,
			Comment: it.ColumnComment,
		}

		if field.Null == "YES" {
			field.Nullable = true
		}

		field.GoType = t.getGoType(field.Type)

		fields = append(fields, field)
	}

	return
}

func (t *mysql) getGoType(dbType string) string {
	// 精确匹配
	if v, ok := typeMysqlDic[dbType]; ok {
		return v
	}

	// 正则匹配
	for _, v := range typeMysqlMatch {
		if ok, _ := regexp.MatchString(v[0], dbType); ok {
			return v[1]
		}
	}

	panic(fmt.Sprintf("unkonow type: %s", dbType))
}

// TypeMysqlDicMp Accurate matching type.精确匹配类型
var typeMysqlDic = map[string]string{
	"smallint":            "int16",
	"smallint unsigned":   "uint16",
	"int":                 "int",
	"int unsigned":        "uint",
	"bigint":              "int64",
	"bigint unsigned":     "uint64",
	"varchar":             "string",
	"char":                "string",
	"date":                "time.Time",
	"datetime":            "time.Time",
	"bit(1)":              "int8",
	"tinyint":             "int8",
	"tinyint unsigned":    "uint8",
	"tinyint(1)":          "int8",
	"tinyint(1) unsigned": "uint8",
	"json":                "string",
	"text":                "string",
	"timestamp":           "time.Time",
	"double":              "float64",
	"mediumtext":          "string",
	"longtext":            "string",
	"float":               "float32",
	"tinytext":            "string",
	"enum":                "string",
	"time":                "time.Time",
	"blob":                "[]byte",
	"tinyblob":            "[]byte",
}

// TypeMysqlMatchMp Fuzzy Matching Types.模糊匹配类型
var typeMysqlMatch = [][]string{
	{`^(tinyint)[(]\d+[)] unsigned`, "uint8"},
	{`^(tinyint)[(]\d+[)]`, "int8"},
	{`^(smallint)[(]\d+[)]`, "int16"},
	{`^(int)[(]\d+[)]`, "int"},
	{`^(bigint)[(]\d+[)] unsigned`, "uint64"},
	{`^(bigint)[(]\d+[)]`, "int64"},
	{`^(char)[(]\d+[)]`, "string"},
	{`^(enum)[(](.)+[)]`, "string"},
	{`^(set)[(](.)+[)]`, "string"},
	{`^(varchar)[(]\d+[)]`, "string"},
	{`^(varbinary)[(]\d+[)]`, "[]byte"},
	{`^(binary)[(]\d+[)]`, "[]byte"},
	{`^(tinyblob)[(]\d+[)]`, "[]byte"},
	{`^(decimal)[(]\d+,\d+[)]`, "float64"},
	{`^(mediumint)[(]\d+[)]`, "string"},
	{`^(double)[(]\d+,\d+[)]`, "float64"},
	{`^(float)[(]\d+,\d+[)]`, "float64"},
	{`^(datetime)[(]\d+[)]`, "time.Time"},
	{`^(timestamp)[(]\d+[)]`, "time.Time"},
}
