package handler

import (
	"Alfred/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	rPool "Alfred/cache/redis"
	"strings"
	"time"
	dblayer "Alfred/db"
)

//MultipartUploadInfo：分块信息
type MultipartUploadInfo struct {
	Filehash string
	FileSize int
	UploadID string
	ChunkSize int
	ChunkCount int
}

//初始化分快上传
func InitialMultipartUploadHandler(w http.ResponseWriter,r *http.Request){
	//解析用户请求
	r.ParseForm()
	username:=r.Form.Get("username")
	filehash:=r.Form.Get("filehash")
	filesize,err:=strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1,"parameter invalid",nil).JSONBytes())
		return
	}

	//获得redis链接
	rConn:=rPool.RedisPool().Get()
	defer rConn.Close()


	//生成分块上传初始化信息
	upInfo:=MultipartUploadInfo{
		Filehash:   filehash,
		FileSize:   filesize,
		UploadID:   username+fmt.Sprintf("%x",time.Now().UnixNano()),
		ChunkSize:  5*1024*1024,//5MB
		ChunkCount: int(math.Ceil(float64(filesize)/(5*1024*1024))),
	}


	//将信息写入redis
	rConn.Do("HSET","MP_"+upInfo.UploadID,"chunkcount",upInfo.ChunkCount)
	rConn.Do("HSET","MP_"+upInfo.UploadID,"filehash",upInfo.Filehash)
	rConn.Do("HSET","MP_"+upInfo.UploadID,"filesize",upInfo.FileSize)

	//将响应初始化写到客户端
	w.Write(util.NewRespMsg(0,"OK",upInfo).JSONBytes())

}

// UploadPartHandler : 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	//	username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	// 2. 获得redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 获得文件句柄，用于存储分块内容
	fpath := "/tmp/" + uploadID + "/" + chunkIndex
	fmt.Print(fpath)
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 4. 更新redis缓存状态
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

//CompleteUploadHandler：通知上传合并
func CompleteUploadHandler(w http.ResponseWriter,r *http.Request){
	//解析用户请求
	r.ParseForm()
	upid:=r.Form.Get("uploadID")
	username:=r.Form.Get("username")
	filehash:=r.Form.Get("filehash")
	filesize:=r.Form.Get("filesize")
	filename:=r.Form.Get("filename")

	//获得redis链接
	rConn:=rPool.RedisPool().Get()
	defer rConn.Close()

	//通过uploadid查询是否完成所有分快的上传
	data,err:=redis.Values(rConn.Do("HGETALL","MP"+upid))
	if err!=nil {
		w.Write(util.NewRespMsg(-1,"complete upload failed",nil).JSONBytes())
		return
	}
	totalcount:=0
	chunkcount:=0

	for i:=0;i<len(data);i+=2 {
		k:=string(data[i].([]byte))
		v:=string(data[i+1].([]byte))
		if k=="chunkcount" {
			totalcount,_=strconv.Atoi(v)
		}else if strings.HasPrefix(k,"chkidx_")&&v=="1"{
			chunkcount++
		}
	}
	if totalcount!=chunkcount{
		w.Write(util.NewRespMsg(-2,"invalid request",nil).JSONBytes())
		return
	}
	//分块合并操作

	//更新唯一文件表和用户文件表
	fsize, _ := strconv.Atoi(filesize)
	dblayer.UploadMetaToDB(filehash, filename, int64(fsize), "")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	//向客户端响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}
