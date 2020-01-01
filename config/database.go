//Package config ...
package config

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

func (app *App) DB() *sql.DB {
	db, err := sql.Open("mysql", app.DBUserName+":"+app.DBPass+"@/"+app.DBName)
	if err != nil {
		log.Fatalln(err)
	}
	db.SetConnMaxLifetime(time.Second * 5)
	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(151)
	return db
}
