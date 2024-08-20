package database

import (
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"
)

func Connect(host, dbname, user, passwd string) *sql.DB {
	cfg := mysql.Config{
		User:                 user,
		Passwd:               passwd,
		Net:                  "tcp",
		Addr:                 host,
		DBName:               dbname,
		Loc:                  time.UTC,
		MaxAllowedPacket:     64 << 20, // same as mysql.defaultMaxAllowedPacket
		ParseTime:            true,
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return db
}
