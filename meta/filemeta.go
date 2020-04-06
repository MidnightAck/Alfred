package meta

type FileMeta struct {
	Filename string
	FileHash string
	Location string
	UploadAt string
	FileSize int64
}

var filemetas map[string]FileMeta

//初始化FileMeta
func init(){
	filemetas = make(map[string]FileMeta)
}

func UpdateFileMeta(fmeta FileMeta){
	filemetas[fmeta.FileHash]=fmeta
}

func GetFileMeta(filesha1 string) FileMeta{
	return filemetas[filesha1]
}

func RemoveFileMeta(filesha1 string){
	delete(filemetas,filesha1)
}
