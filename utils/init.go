package utils

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

var Config cfg
var Database *sql.DB
var Log *logrus.Logger

func LoadUtils() error {
	var conf = loadConfig()
	var db, err = openDB(conf.DSN)

	if err != nil {
		fmt.Println("Could not connect to db.", err)
		return err
	}

	Config = conf
	Database = db
	Log = loadLogger()
	return nil
}
