package handler

import (
	cmn "Alfred/common"
	cfg "Alfred/config"
	"Alfred/meta"
	"Alfred/mq"
	"Alfred/store/oss"
	"Alfred/util"
	"encoding/json"
	"fmt"
	//"gopkg.in/amz.v1/s3"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	dblayer "Alfred/db"
	"Alfred/store/ceph"
)

//UploadHandler：处理文件上传
func UploadHandler(w http.ResponseWriter,r *http.Request){
	//判断请求类型
	if r.Method=="GET"{
		//GET方法直接返回html
		data,err:=ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w,"internal server error")
			return
		}
		io.WriteString(w,string(data))
	}else if r.Method=="POST"{
		//接受文件流并且存储到本地
		data,head,err:=r.FormFile("file")	//key应当与html中的id保持一致
		if err != nil {
			io.WriteString(w,"internal server error")
			return
		}
		defer data.Close()

		fmeta:=meta.FileMeta{
			Filename: head.Filename,
			//FileHash: "",
			Location: "/tmp/"+head.Filename,
			UploadAt: time.Now().Format("2020-04-06 12:45:00"),
			//FileSize: 0,
		}

		//申请本地空间
		newfile,err:=os.Create(fmeta.Location)
		if err != nil {
			io.WriteString(w,"Failed to create space err:"+err.Error())
			return
		}
		defer  newfile.Close()

		//将文件从内存复制到本地
		fmeta.FileSize,err=io.Copy(newfile,data)
		if err != nil {
			io.WriteString(w,"Failed to save data into file err:"+err.Error())
			return
		}

		//在meta中插入信息
		newfile.Seek(0,0)
		fmeta.FileHash=util.FileSha1(newfile)
		//meta.UpdateFileMeta(fmeta)

		// 5. 判断之前是否上传过，如果上传过，则只在用户文件表添加记录


		//将文件写入云端


		if cfg.CurrentStoreType == cmn.StoreCeph {
			// 文件写入Ceph存储
			data, _ := ioutil.ReadAll(newfile)
			cephPath := "/ceph/" + fmeta.FileHash
			_ = ceph.PutObject("userfile", cephPath, data)
			fmeta.Location = cephPath
		} else if cfg.CurrentStoreType == cmn.StoreOSS {
			// 文件写入OSS存储
			ossPath := "oss/" + fmeta.FileHash
			// 判断写入OSS为同步还是异步
			if !cfg.AsyncTransferEnable {
				err = oss.Bucket().PutObject(ossPath, newfile)
				if err != nil {
					fmt.Println(err.Error())
					w.Write([]byte("Upload failed!"))
					return
				}
				fmeta.Location = ossPath
			} else {
				// 写入异步转移任务队列
				data := mq.TransferData{
					FileHash:      fmeta.FileHash,
					CurLocation:   fmeta.Location,
					DestLocation:  ossPath,
					DestStoreType: cmn.StoreOSS,
				}
				pubData, _ := json.Marshal(data)
				pubSuc := mq.Publish(
					cfg.TransExchangeName,
					cfg.TransOSSRoutingKey,
					pubData,
				)
				if !pubSuc {
					// TODO: 当前发送转移信息失败，稍后重试
				}
			}
		}


		meta.UploadFileMetaDB(fmeta)

		//更新用户文件表
		r.ParseForm()
		username:=r.Form.Get("username")
		suc:=dblayer.OnUserFileUploadFinished(username,fmeta.FileHash,fmeta.Filename,fmeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		}else{
			w.Write([]byte("upload failed"))
		}
		//重定向到成功页面
		//http.Redirect(w,r,"/file/upload/suc",http.StatusFound)
	}
}

//UploadSuc:上传成功后打印信息
func UploadSuc(w http.ResponseWriter,r *http.Request){
	io.WriteString(w,"Upload Succeed!")
}

//DownloadHandler:下载文件
func DownloadHandler(w http.ResponseWriter,r *http.Request){
	//解析命令
	r.ParseForm()
	fsha1:=r.Form.Get("filehash")

	fmeta,_:=meta.GetFileMetaDB(fsha1)

	//os打开该位置的文件
	f,err:=os.Open(fmeta.Location)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	//读取打开的文件
	data,err:=ioutil.ReadAll(f)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//开启下载窗口
	w.Header().Set("Content-Type","application/octect-stream")
	w.Header().Set("Content-Description","attachment;filename=\""+fmeta.Filename+"\"")
	w.Write(data)
}

//DownloadFile：提供下载文件的接口
func DownloadFile(w http.ResponseWriter,loc string, filename string){
	//os打开该位置的文件
	f,err:=os.Open(loc)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	//读取打开的文件
	data,err:=ioutil.ReadAll(f)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//开启下载窗口
	w.Header().Set("Content-Type","application/octect-stream")
	w.Header().Set("Content-Description","attachment;filename=\""+filename+"\"")
	w.Write(data)
}

//GetFilemeta:获取文件元信息
func GetFilemeta(w http.ResponseWriter,r *http.Request){
	r.ParseForm()
	fsha1:=r.Form.Get("filehash")
	//fmeta:=meta.GetFileMeta(fsha1)
	fmeta,err:=meta.GetFileMetaDB(fsha1)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data,err:=json.Marshal(fmeta)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// FileQueryHandler : 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	//fileMetas, _ := meta.GetLastFileMetasDB(limitCnt)
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

//UpdateFileMeta:更新文件元信息，此处只可更改文件名称
func UpdateFileMeta(w http.ResponseWriter,r *http.Request){
	r.ParseForm()
	fsha1:=r.Form.Get("filehash")
	fname:=r.Form.Get("filename")

/*
	if r.Method !="POST"{
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Printf("500")
		return
	}
*/

	fmeta,err:=meta.GetFileMetaDB(fsha1)
	fmeta.Filename=fname
	meta.UpdateFileMetaDB(fmeta)

	data,err:=json.Marshal(fmeta)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//FileDeleteHandler:删除文件以及文件元信息
func FileDeleteHandler(w http.ResponseWriter,r *http.Request){
	r.ParseForm()
	filsha1:=r.Form.Get("filehash")

	fMeta:=meta.GetFileMeta(filsha1)
	os.Remove(fMeta.Location)

	//meta.RemoveFileMeta(filsha1)
	meta.RemoveFileMetaDB(filsha1)
	w.WriteHeader(http.StatusOK)
}

// TryFastUploadHandler : 尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1. 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4. 上传过则将文件信息写入用户文件表， 返回成功
	suc := dblayer.OnUserFileUploadFinished(
		username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}


// DownloadURLHandler : 生成文件的下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	filehash := r.Form.Get("filehash")
	// 从文件表查找记录
	row, _ := dblayer.GetMetaFromDB(filehash)//meta.GetFileMetaDB(filehash)

	// TODO: ceph 速度较慢可以优化
	if strings.HasPrefix(row.FileAddr.String, "/tmp") {
		username := r.Form.Get("username")
		token := r.Form.Get("token")
		tmpUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			r.Host, filehash, username, token)
		w.Write([]byte(tmpUrl))
	}else if strings.HasPrefix(row.FileAddr.String, "/ceph") {
		bucket := ceph.GetCephBucket("userfile")
		path:="/ceph"+row.FileHash
		d, _ := bucket.Get(path)
		//fmt.Print(path)
		//w.Write([]byte(d))
		tmpFile, _ := os.Create("/tmp/test_file")
		tmpFile.Write(d)
	}  else if strings.HasPrefix(row.FileAddr.String, "oss/") {
		// oss下载url
		signedURL := oss.DownloadURL(row.FileAddr.String)
		fmt.Print(signedURL)
		w.Write([]byte(signedURL))
	}

}

