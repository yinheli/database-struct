package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

func newDb(dialect, dsn string) (db *gorm.DB, err error) {
	db, err = gorm.Open(dialect, dsn)
	if err != nil {
		return
	}

	db.DB().SetMaxOpenConns(10)
	db.DB().SetMaxIdleConns(0)
	db.DB().SetConnMaxLifetime(time.Minute * 5)
	db.SingularTable(true)

	err = db.DB().Ping()
	return
}
