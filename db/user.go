package db

import (
	mydb "Alfred/db/mysql"
	"fmt"
)

func UserSignUp(username string,password string) bool {
	stmt,err:=mydb.DBConn().Prepare("insert into tbl_user (`user_name`,`user_pwd`) values (?,?)")
	if err!=nil {
		fmt.Printf("Failed to prepare sql err:"+err.Error())
		return false
	}
	defer stmt.Close()

	res,err:=stmt.Exec(username,password)
	if err!=nil {
		fmt.Printf("Failed to excute sql err:"+err.Error())
		return false
	}

	if rf,err:=res.RowsAffected();err==nil {
		if rf<=0{
			fmt.Printf("user with username:"+username+"has already existed")
			return false
		}
	}
	return true
}

func UserSignIn(username string,password string) bool {

}