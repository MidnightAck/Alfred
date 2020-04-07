package db

import (
	mydb "Alfred/db/mysql"
	"fmt"
)

func UploadFileToDB(filehash string,filename string,filesize int64,fileaddr string) bool {
	stmt,err:=mydb.DBConn().Prepare("insert into tbl_file (`file_sha1`,`file_name`,`file_size`," +
		"`file_addr`,`status`) values (?,?,?,?,1)")
	if err !=nil {
		fmt.Printf("Failed to prepare sql,err:"+err.Error())
		return false
	}
	defer stmt.Close()

	res,err:=stmt.Exec(filehash,filename,filesize,fileaddr)
	if err !=nil {
		fmt.Printf("Failed to excute sql,err:"+err.Error())
		return false
	}

	if rf,err:=res.RowsAffected();err==nil {
		if rf<=0 {
			fmt.Println("File with filehash"+filehash+"has already been uploaded")
		}
		return true
	}
	return false
}