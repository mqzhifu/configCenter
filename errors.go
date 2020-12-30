package configCenter

import (
	"encoding/json"
	"strings"
	"fmt"
	"path/filepath"
	"runtime"
)

type ErrorCode struct {
	Code  uint32
	Msg   string
	Where string
}

func (e *ErrorCode) Error() string {
	//var errorCode ErrorCode {}
	errorCode := new (ErrorCode)
	errorCode.Code = e.Code
	errorCode.Msg = e.Msg
	errorCode.Where = e.Where

	errorJsonStr,_ := json.Marshal(errorCode)
	return string(errorJsonStr)
}
// 声明一个错误
func NewCoder(code uint32, msg string) *ErrorCode {
	where := caller(1, false)
	return &ErrorCode{Code: code, Msg: msg, Where: where}
}

// 对一个错误追加信息
func Wrap(err error, extMsg ...string) *ErrorCode {
	msg  := err.Error()
	if len(extMsg) != 0 {
		msg = strings.Join(extMsg, " : ") + " : " + msg
	}
	return &ErrorCode{Msg: msg}
}

// 获取源代码行数
func caller(calldepth int, short bool) string {
	_, file, line, ok := runtime.Caller(calldepth + 1)
	if !ok {
		file = "???"
		line = 0
	} else if short {
		file = filepath.Base(file)
	}

	return fmt.Sprintf("%s:%d", file, line)
}

func myPrint(a ...interface{}) (n int, err error) {
	return fmt.Println(a)
}

