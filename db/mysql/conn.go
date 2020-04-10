package mysql
//提供数据库连接的功能

import (
	cfg "Alfred/config"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql-1.5.0"
	"log"
	"os"
)

var db *sql.DB


func init(){
	db, _ = sql.Open("mysql", cfg.MySQLSource)
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

func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		//将行数据保存到record字典
		err := rows.Scan(scanArgs...)
		checkErr(err)

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		records = append(records, record)
	}
	return records
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
