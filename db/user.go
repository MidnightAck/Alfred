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
	stmt,err:=mydb.DBConn().Prepare("select * from tbl_user where user_name=? ")
	if err!=nil {
		fmt.Printf("Failed to prepare sql err:"+err.Error())
		return false
	}
	defer stmt.Close()

	res,err:=stmt.Query(username)
	if err!=nil {
		fmt.Printf("Fail to excute select err:"+err.Error())
		return false
	}else if res==nil{
		fmt.Printf("username not found:"+username)
		return false
	}

	pRows:=mydb.ParseRows(res)
	if len(pRows)>0 && string(pRows[0]["user_pwd"].([]byte))==password{
		return true
	}
	return false

}

//刷新用户登陆的token
func UpdateToken(username string,token string) bool {
	stmt,err:=mydb.DBConn().Prepare("replace into tbl_user_token (`user_name`,`user_token`) values (?,?)")
	if err!=nil{
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_,err =stmt.Exec(username,token)
	if err!=nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

type User struct{
	Username string
	Email string
	Phone string
	SignupAt string
	LastActiveAt string
	Status int
}

func GetUserInfo(username string) (User,error){
	user:=User{}

	stmt,err:=mydb.DBConn().Prepare("select user_name,signup_at from tbl_user where user_name=? limit 1")
	if err!=nil{
		fmt.Println(err.Error())
		return user,err
	}
	defer stmt.Close()

	err=stmt.QueryRow(username).Scan(&user.Username,&user.SignupAt)
	if err!=nil{
		fmt.Println(err.Error())
		return user,err
	}

	return user,nil
}