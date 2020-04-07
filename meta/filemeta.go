package meta

import (
	mydb "Alfred/db"
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
	return mydb.UploadFileToDB(fmeta.FileHash,fmeta.Filename,fmeta.FileSize,fmeta.Location)
}

//GetFileMeta：根据sha1查找FileMeta
func GetFileMeta(filesha1 string) FileMeta{
	return filemetas[filesha1]
}

//GetFileMetaDB：根据sha1从数据库中查找FileMeta
func GetFileMetaDB(filesha1 string) (FileMeta,error){
	tfile,err:=mydb.GetMetaFromDB(filesha1)
	if err!=nil {
		return FileMeta{},err
	}
	fmeta:=FileMeta{
		FileHash: tfile.FileHash,
		Filename: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
		UploadAt: "",
	}
	return fmeta,nil
}

//RemoveFileMeta：根据key删除FileMeta
func RemoveFileMeta(filesha1 string){
	delete(filemetas,filesha1)
}

