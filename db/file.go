package db
//负责数据库中的文件操作，提供增删改查的功能实现

import (
	mydb "Alfred/db/mysql"
	"database/sql"
	"fmt"
)

//UploadFileToDB:完成文件向数据库上传的功能
func UploadMetaToDB(filehash string,filename string,filesize int64,fileaddr string) bool {
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

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
	CreateAT string
}

//GetMetaFromDB：获取文件元信息
func GetMetaFromDB(filehash string) (*TableFile,error) {
	stmt,err:=mydb.DBConn().Prepare("select file_sha1,file_addr,file_name,file_size from " +
		"tbl_file where file_sha1=? and status =1 limit 1")
	if err !=nil {
		fmt.Printf("Failed to prepare sql,err:"+err.Error())
		return nil,err
	}
	defer stmt.Close()

	resfile:=TableFile{}
	//QueryRow根据key返回至多一行，并且可以通过scan方法直接读取
	err=stmt.QueryRow(filehash).Scan(&resfile.FileHash,&resfile.FileAddr,&resfile.FileName,&resfile.FileSize)
	if err !=nil {
		fmt.Printf("Failed to excute sql,err:"+err.Error())
		return nil,err
	}
	return &resfile,nil
}

//UpdateMetaInDB:更新文件元信息名称
func UpdateMetaInDB(filehash string,filename string) bool {
	stmt,err :=mydb.DBConn().Prepare("update tbl_file set file_name=? where file_sha1=? ")
	if err !=nil {
		fmt.Printf("Failed to prepare sql,err:"+err.Error())
		return false
	}
	defer stmt.Close()

	res,err:=stmt.Exec(filename,filehash)
	if err !=nil {
		fmt.Printf("Failed to excute sql,err:"+err.Error())
		return false
	}

	if rf,err:=res.RowsAffected();err==nil {
		if rf<=0 {
			fmt.Println("there is something wrong")
			return false
		}
		return true
	}
	return false
}

//DeleteFileFromDB:从数据库中删除Meta
func DeleteFileFromDB(filehash string) bool {
	stmt,err:=mydb.DBConn().Prepare("delete from tbl_file where file_sha1=?")
	if err !=nil {
		fmt.Printf("Failed to prepare sql,err:"+err.Error())
		return false
	}
	defer stmt.Close()

	res,err:=stmt.Exec(filehash)
	if err !=nil {
		fmt.Printf("Failed to excute sql,err:"+err.Error())
		return false
	}

	if rf,err:=res.RowsAffected();err==nil {
		if rf<=0 {
			fmt.Println("there is something wrong")
			return false
		}
		return true
	}
	return false
}


// GetFileMetaList : 从mysql批量获取文件元信息
func GetFileMetaList(limit int) ([]TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where status=1 limit ?")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	cloumns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(cloumns))
	var tfiles []TableFile
	for i := 0; i < len(values) && rows.Next(); i++ {
		tfile := TableFile{}
		err = rows.Scan(&tfile.FileHash, &tfile.FileAddr,
			&tfile.FileName, &tfile.FileSize)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		tfiles = append(tfiles, tfile)
	}
	fmt.Println(len(tfiles))
	return tfiles, nil
}

// UpdateFileLocation : 更新文件的存储地址(如文件被转移了)
func UpdateFileLocation(filehash string, fileaddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_file set`file_addr`=? where  `file_sha1`=? limit 1")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("更新文件location失败, filehash:%s", filehash)
		}
		return true
	}
	return false
}

// IsFileUploaded : check if hash already exists
func IsFileUploaded(hash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_name, file_size, file_addr, file_sha1, create_at " +
			" from tbl_file where file_sha1 = ? and status = 1 limit 1")
	if err != nil {
		fmt.Println("Failded to prepare statement, err: ", err.Error())
		return nil, err
	}
	tfile := TableFile{}
	err = stmt.QueryRow(hash).Scan(
		&tfile.FileName, &tfile.FileSize, &tfile.FileAddr, &tfile.FileHash, &tfile.CreateAT)
	if err != nil {
		fmt.Println("IsFileUploaded ", err.Error())
		return nil, err
	}
	return &tfile, nil
}