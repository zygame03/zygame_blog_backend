package httpserver

import "fmt"

const (
	SUCCESS = 0
	FAIL    = 500
)

type Result struct {
	code int
	msg  string
}

func (e Result) Code() int {
	return e.code
}

func (e Result) Msg() string {
	return e.msg
}

var (
	_codes   = map[int]struct{}{}
	_message = make(map[int]string)
)

func RegisterResult(code int, msg string) Result {
	if _, ok := _codes[code]; ok {
		panic(fmt.Sprintf("错误码 %d 已存在", code))
	}
	if msg == "" {
		panic("错误码不能为空")
	}

	_codes[code] = struct{}{}
	_message[code] = msg

	return Result{
		code: code,
		msg:  msg,
	}
}

func GetMsg(code int) string {
	return _message[code]
}

var (
	SuccessResult = RegisterResult(SUCCESS, "SUCCESS")
	FailResult    = RegisterResult(FAIL, "FAIL")
)

var (
	ErrRequest = RegisterResult(1001, "请求参数错误")
	ErrDBOp    = RegisterResult(1002, "数据库操作异常")

	ErrPassword  = RegisterResult(2001, "密码错误")
	ErrUserExist = RegisterResult(2002, "用户不存在")
)
