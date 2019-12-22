//Package config ...
package config

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"time"
)

func DB() (*sql.DB, error) {
	db, err := sql.Open("mysql", viper.GetString("DATABASES.MYSQL.USERNAME")+":"+viper.GetString("DATABASES.MYSQL.PASSWORD")+"@/"+viper.GetString("DATABASES.MYSQL.DATABASE"))
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Second * 5)
	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(151)
	return db, nil
}
