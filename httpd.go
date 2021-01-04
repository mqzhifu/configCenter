package configcenter

import (
	"net/http"
	"strconv"
	"encoding/json"
	"fmt"
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
func NewHttpd(port int ,host string,configer *Configer) *Httpd{
	httpd := new (Httpd)
	httpd.Port = port
	httpd.Host = host
	httpd.configer = configer
	return httpd
}

func (httpd *Httpd)Start(){
	http.HandleFunc("/", httpd.RouterHandler)
	dns := httpd.Host + ":" + strconv.Itoa(httpd.Port)
	myPrint("httpd start loop:",dns)

	err := http.ListenAndServe(dns, nil)
	if err != nil {
		fmt.Printf("http.ListenAndServe()函数执行错误,错误为:%v\n", err)
		return
	}
}

func ResponseMsg(w http.ResponseWriter,code int ,msg interface{} ){
	//fmt.Println("SetResponseMsg in",code,msg)
	responseMsgST := ResponseMsgST{Code:code,Msg:msg}
	//fmt.Println("responseMsg : ",responseMsg)
	jsonResponseMsg , err := json.Marshal(responseMsgST)
	fmt.Println("SetResponseMsg rs",err, string(jsonResponseMsg))

	_, _ = w.Write(jsonResponseMsg)

}

func ResponseStatusCode(w http.ResponseWriter,code int ,responseInfo string){
	w.Header().Set("Content-Length",strconv.Itoa( len(responseInfo) ) )
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(403)
	w.Write([]byte(responseInfo))
}

func (httpd *Httpd)RouterHandler(w http.ResponseWriter, r *http.Request){
	//fmt.Printf("时间戳（秒）：%v;\n", time.Now().Unix())
	//time.Now().UnixNano()

	myPrint("receiver: have a new request.(", time.Now().Format("2006-01-02 15:04:05"),")")
	parameter := r.URL.Query()
	fmt.Println("url.query",parameter)
	fmt.Println("uri",r.URL.RequestURI())
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
		ResponseStatusCode(w,500,"RequestURI is null")
		return
	}

	searchRs,_ := httpd.configer.Search(r.URL.RequestURI())
	ResponseMsg(w,200,searchRs)

}


