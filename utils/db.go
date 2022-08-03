package utils

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return db, nil
}
