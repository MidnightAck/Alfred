package mysql

import (
	"database/sql"
	"fmt"
	"os"
	_ "github.com/go-sql-driver/mysql-1.5.0"
	"strings"
)

var db *sql.DB

const(
	userName="Alfred"
	password="123456"
	ip="192.144.155.134"
	port="3306"
	dbName="Alfred"
)

func init(){
	path := strings.Join([]string{userName, ":", password, "@tcp(",ip, ":", port, ")/", dbName, "?charset=utf8"}, "")
	db,_=sql.Open("mysql",path)
	db.SetMaxIdleConns(1000)
	err:=db.Ping()
	if err!=nil {
		fmt.Printf("Failed to connect to db err:"+err.Error())
		os.Exit(1)
	}

}

func DBConn() *sql.DB {
	return db
}
