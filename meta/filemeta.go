package meta

import (
	mydb "Alfred/db"
	"fmt"
)

//文件元信息的数据结构，包含
//Filename 文件的名称
//FileHash 文件的hash值，为唯一标志
//Location 文件存储的位置
//UploadAt 文件上传时间
//FileSize 文件大小
type FileMeta struct {
	Filename string
	FileHash string
	Location string
	UploadAt string
	FileSize int64
}

//全局的filemeta，key为文件的hash
var filemetas map[string]FileMeta

//init：初始化FileMeta
func init(){
	filemetas = make(map[string]FileMeta)
}

//UpdateFileMeta：更新或插入filemeta
func UpdateFileMeta(fmeta FileMeta){
	filemetas[fmeta.FileHash]=fmeta
}

//UploadFileMetaDB:向数据库插入filemeta
func UploadFileMetaDB(fmeta FileMeta) bool {
	return mydb.UploadMetaToDB(fmeta.FileHash,fmeta.Filename,fmeta.FileSize,fmeta.Location)
}

//UploadFileMetaDB:向数据库更新Meta的名称
func UpdateFileMetaDB(fmeta *FileMeta) bool {
	res:= mydb.UpdateMetaInDB(fmeta.FileHash,fmeta.Filename)
	if res==false {
		fmt.Printf("Not Changed")
	}
	return res
}

//GetFileMeta：根据sha1查找FileMeta
func GetFileMeta(filesha1 string) FileMeta{
	return filemetas[filesha1]
}

//GetFileMetaDB：根据sha1从数据库中查找FileMeta
func GetFileMetaDB(filesha1 string) (*FileMeta,error){
	tfile,err:=mydb.GetMetaFromDB(filesha1)
	if err!=nil {
		return nil,err
	}
	fmeta:=FileMeta{
		FileHash: tfile.FileHash,
		Filename: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
		UploadAt: "",
	}
	return &fmeta,nil
}

//RemoveFileMeta：根据key删除FileMeta
func RemoveFileMeta(filesha1 string){
	delete(filemetas,filesha1)
}

//RemoveFileMetaDB：根据key删除FileMeta
func RemoveFileMetaDB(filesha1 string) bool {
	return mydb.DeleteFileFromDB(filesha1)
}

// IsFileUploadedDB : check if file has checked
func IsFileUploadedDB(hash string) (FileMeta, error) {
	tfile, err := mydb.IsFileUploaded(hash)
	if err != nil {
		fmt.Println("IsFileUploadedDB : err:", err.Error())
		return FileMeta{}, err
	}
	fileMeta := FileMeta{
		Filename: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
		FileHash:     tfile.FileHash}
	return fileMeta, nil
}