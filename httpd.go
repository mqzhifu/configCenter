package configcenter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"src/zlib"
	"strconv"
	"strings"
	"time"
)

type ResponseMsgST struct {
	Code 	int
	Msg 	interface{}
}

type Httpd struct {
	Port int
	Host string
	configer *Configer

}

func NewHttpd(port int ,host string,configerInitData Configer) *Httpd{
	zlib.MyPrint("NewHttpd : ",host,port,configerInitData)
	httpd := new (Httpd)
	httpd.Port = port
	httpd.Host = host
	configer := NewConfiger(
		configerInitData.FileTotalSizeMax ,
		configerInitData.FileSizeMax,
		configerInitData.FileCntMax,
		configerInitData.AllowExtType,
		)

	err := configer.StartLoading(configerInitData.RootPath)
	if err != nil{
		zlib.MyPrint("StartLoading err",err.Error())
	}

	httpd.configer = configer
	return httpd
}

func (httpd *Httpd)Start(){

	http.HandleFunc("/", httpd.RouterHandler)
	dns := httpd.Host + ":" + strconv.Itoa(httpd.Port)
	zlib.MyPrint("httpd start loop:",dns)

	err := http.ListenAndServe(dns, nil)
	if err != nil {
		fmt.Printf("http.ListenAndServe()函数执行错误,错误为:%v\n", err)
		return
	}
}

func ResponseMsg(w http.ResponseWriter,code int ,msg string ){
	//fmt.Println("SetResponseMsg in",code,msg)
	//responseMsgST := ResponseMsgST{Code:code,Msg:msg}
	responseMsgST := ResponseMsgST{Code:code,Msg:"#msg#"}
	msg = msg[1:len(msg)-1]
	//这里有个无奈的地方，为了兼容非网络请求，正常使用时，返回的就是json,现在HTTP套一层，还得再一层JSON，冲突了
	//fmt.Println("responseMsg : ",responseMsg)
	jsonResponseMsg , err := json.Marshal(responseMsgST)
	jsonResponseMsgNew := strings.Replace(string(jsonResponseMsg),"#msg#",msg,-1)
	fmt.Println("SetResponseMsg rs",err, string(jsonResponseMsgNew))

	_, _ = w.Write([]byte(jsonResponseMsgNew))

}

func ResponseStatusCode(w http.ResponseWriter,code int ,responseInfo string){
	w.Header().Set("Content-Length",strconv.Itoa( len(responseInfo) ) )
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(403)
	w.Write([]byte(responseInfo))
}
//主要，是接收HTTP 回调
func (httpd *Httpd)RouterHandler(w http.ResponseWriter, r *http.Request){
	//fmt.Printf("时间戳（秒）：%v;\n", time.Now().Unix())
	//time.Now().UnixNano()

	zlib.MyPrint("receiver: have a new request.(", time.Now().Format("2006-01-02 15:04:05"),")")
	parameter := r.URL.Query()
	zlib.MyPrint("uri",r.URL.RequestURI(),"url.query",parameter)
	if r.URL.RequestURI() == "/favicon.ico" {
		ResponseStatusCode(w,403,"no power")
		return
	}
	//appName := parameter.Get("app_name")
	//fmt.Println("app_name",appName)
	//if appName == "" {
	//	ResponseMsg(w,500,"appName is null")
	//	return
	//}
	if r.URL.RequestURI() == "" || r.URL.RequestURI() == "/" {
		ResponseStatusCode(w,500,"RequestURI is null or not  '/' start")
		return
	}

	searchRs,err := httpd.configer.Search(r.URL.RequestURI())
	if err != nil {
		zlib.MyPrint("search err",err.Error())
		ResponseMsg(w,500,err.Error())
	}
	zlib.MyPrint("searchRs",searchRs)
	ResponseMsg(w,200,searchRs)

}


