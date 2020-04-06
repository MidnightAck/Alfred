package handler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"Alfred/meta"
	"time"
	"Alfred/util"
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

		//在meta中更新信息
		newfile.Seek(0,0)
		fmeta.FileHash=util.FileSha1(newfile)
		meta.UpdateFileMeta(fmeta)

		//重定向到成功页面
		http.Redirect(w,r,"/file/upload/suc",http.StatusFound)
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

	fmeta:=meta.GetFileMeta(fsha1)

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

//GetFilemeta:获取文件元信息
func GetFilemeta(w http.ResponseWriter,r *http.Request){
	r.ParseForm()
	fsha1:=r.Form.Get("filehash")
	fmeta:=meta.GetFileMeta(fsha1)

	data,err:=json.Marshal(fmeta)
	if err!=nil {
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

	if r.Method !="POST"{
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fmeta:=meta.GetFileMeta(fsha1)
	fmeta.Filename=fname
	meta.UpdateFileMeta(fmeta)

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

	meta.RemoveFileMeta(filsha1)
	w.WriteHeader(http.StatusOK)
}