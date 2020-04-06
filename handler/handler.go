package handler

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

//处理文件上传
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

		newfile,err:=os.Create("/tmp/"+head.Filename)
		if err != nil {
			io.WriteString(w,"Failed to create space err:"+err.Error())
			return
		}
		defer  newfile.Close()

		_,err=io.Copy(newfile,data)
		if err != nil {
			io.WriteString(w,"Failed to save data into file err:"+err.Error())
			return
		}

		http.Redirect(w,r,"/file/upload/suc",http.StatusFound)
	}
}

//上传成功后打印信息
func UploadSuc(w http.ResponseWriter,r *http.Request){
	io.WriteString(w,"Upload Succeed!")
}