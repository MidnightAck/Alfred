package handler

import (
	rPool "Alfred/cache/redis"
	dblayer "Alfred/db"
	"Alfred/util"
	"bufio"
	"bytes"
	"cloudstore/config"
	"fmt"
	"github.com/garyburd/redigo/redis"
	jsonit "github.com/json-iterator"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
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
	// 6. 合并分块为一个单独的文件，本质上使用的是如下命令
	// cat `ls | sort -n` > /tmp/filename

	partFileStorePath := config.DirPath + "/uploadid" // 分块所在的目录
	fileStorePath := config.DirPath + filename        // 最后文件保存的路径
	if _, err := mergeAllPartFile(partFileStorePath, fileStorePath); err != nil {
		w.Write(util.NewRespMsg(-2,"failed to merge multi-part files",nil).JSONBytes())
		return
	}


	//更新唯一文件表和用户文件表
	fsize, _ := strconv.Atoi(filesize)
	dblayer.UploadMetaToDB(filehash, filename, int64(fsize), "")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	//向客户端响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// mergeAllPartFile: filepath： 分块存储的路径 filestore： 文件最终地址
func mergeAllPartFile(partFileStorePath, fileStorePath string) (bool, error) {
	var cmd *exec.Cmd
	cmd = exec.Command(config.MergeAllShell, partFileStorePath, fileStorePath)

	if _, err := cmd.Output(); err != nil {
		fmt.Println(err)
		return false, err
	}
	fmt.Println(fileStorePath, " has been merge complete")
	return true, nil
}

func multipartUpload(filename string, targetURL string, chunkSize int) error {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()

	bfRd := bufio.NewReader(f)
	index := 0

	ch := make(chan int)
	buf := make([]byte, chunkSize) //每次读取chunkSize大小的内容
	for {
		n, err := bfRd.Read(buf)
		if n <= 0 {
			break
		}
		index++

		bufCopied := make([]byte, 5*1048576)
		copy(bufCopied, buf)

		go func(b []byte, curIdx int) {
			fmt.Printf("upload_size: %d\n", len(b))

			resp, err := http.Post(
				targetURL+"&index="+strconv.Itoa(curIdx),
				"multipart/form-data",
				bytes.NewReader(b))
			if err != nil {
				fmt.Println(err)
			}

			body, er := ioutil.ReadAll(resp.Body)
			fmt.Printf("%+v %+v\n", string(body), er)
			resp.Body.Close()

			ch <- curIdx
		}(bufCopied[:n], index)

		//遇到任何错误立即返回，并忽略 EOF 错误信息
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err.Error())
			}
		}
	}

	for idx := 0; idx < index; idx++ {
		select {
		case res := <-ch:
			fmt.Println(res)
		}
	}

	return nil
}

func DoMpupload(username string,token string,filehash string,filesize string,filename string) {
	// 1. 请求初始化分块上传接口
	resp, err := http.PostForm(
		"http://localhost:8080/file/mpupload/init",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {filesize},
		})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	// 2. 得到uploadID以及服务端指定的分块大小chunkSize
	uploadID := jsonit.Get(body, "data").Get("UploadID").ToString()
	chunkSize := jsonit.Get(body, "data").Get("ChunkSize").ToInt()
	//fmt.Printf("uploadid: %s  chunksize: %d\n", uploadID, chunkSize)

	// 3. 请求分块上传接口
	tURL := "http://localhost:8080/file/mpupload/uppart?" +
		"username=admin&token=" + token + "&uploadid=" + uploadID
	multipartUpload(filename, tURL, chunkSize)

	// 4. 请求分块完成接口
	resp, err = http.PostForm(
		"http://localhost:8080/file/mpupload/complete",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {filesize},
			"filename": {filename},
			"uploadid": {uploadID},
		})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	fmt.Printf("complete result: %s\n", string(body))
}