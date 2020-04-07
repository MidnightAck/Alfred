package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	dblayer "Alfred/db"
	"Alfred/util"
	"time"
)

const pwd_salt="#55&"

func SignUpHandler(w http.ResponseWriter,r *http.Request){
	if r.Method == http.MethodGet{
		data,err:=ioutil.ReadFile("./static/view/register.html")
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

func UserSignInHandler(w http.ResponseWriter,r *http.Request) {
	r.ParseForm()
	username:=r.Form.Get("username")
	password:=r.Form.Get("password")

	encPasswd:=util.Sha1([]byte(password+pwd_salt))
	check:=dblayer.UserSignIn(username,encPasswd)

	if check==false{
		w.Write([]byte("FAILED"))
		return
	}
	//生成访问凭证 token
	token:=GenToken(username)
	upRes:=dblayer.UpdateToken(username,token)
	if upRes==false {
		w.Write([]byte("FAILED"))
		return
	}
	//登陆成功后重定向到首页
	//w.Write([]byte("http://"+r.Host+"/static/view/home.html"))
	resp:=util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token string
		}{
			Location:"http://"+r.Host+"/static/view/home.html",
			Username:username,
			Token:token,
		},
	}
	w.Write(resp.JSONBytes())


}

func GenToken(username string) string {
	//md5(username+timestamp+token_salt)+timestamp[:8]
	ts:=fmt.Sprintf("%x",time.Now().Unix())
	tokenPrefix:=util.MD5([]byte(username+ts+"_tokensalt"))
	return tokenPrefix + ts[:8]
}

func UserInfoHandler(w http.ResponseWriter,r *http.Request){
	//解析请求参数
	r.ParseForm()
	username :=r.Form.Get("username")
	token:=r.Form.Get("token")
	//验证token是否有效
	isValid:=IsTokenValid(token)
	if !isValid{
		w.WriteHeader(403)
		return
	}
	//查询用户信息
	user,err:=dblayer.GetUserInfo(username)
	if err!=nil {
		w.WriteHeader(403)
		return
	}
	//组装并且响应用户数据
	resp:=util.RespMsg{
		Code:0,
		Msg:"OK",
		Data:user,
	}
	w.Write(resp.JSONBytes())
}


// IsTokenValid : token是否有效
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token的时效性，是否过期
	// TODO: 从数据库表tbl_user_token查询username对应的token信息
	// TODO: 对比两个token是否一致
	return true
}
