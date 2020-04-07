package handler

import (
	"io/ioutil"
	"net/http"
	dblayer "Alfred/db"
	"Alfred/util"
)

const pwd_salt="#55&"

func SignUpHandler(w http.ResponseWriter,r *http.Request){
	if r.Method == http.MethodGet{
		data,err:=ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(data)
		return
	}

	r.ParseForm()
	username:=r.Form.Get("username")
	password:=r.Form.Get("password")

	if valid:=JudgeNameValid(username,password);valid==false{
		w.Write([]byte("Invaild parameter"))
		return
	}

	enc_passwd:=util.Sha1([]byte(password+pwd_salt))
	suc:=dblayer.UserSignUp(username,enc_passwd)
	if suc==true {
		w.Write([]byte("SUCCESS"))
	}else{
		w.Write([]byte("FAILED"))
	}
}

func JudgeNameValid(username string,password string) bool {
	if len(username)<3||len(password)<5 {

		return false
	}
	return true
}