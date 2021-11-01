package response

import "encoding/json"

type Service interface {
	error
	Return() error
	Set(...Option) Service
}

type Options struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
	Msg     string `json:"msg"`
}

type Option func(*Options)

var (
	ErrCodeDefault = 400000 // ErrCodeDefault 默认错误Code
	ErrMsgDefault  = ""     // ErrMsgDefault 默认输出错误信息
	MsgDefault     = ""     // MsgDefault 默认输出信息
)

// ErrCode 定义错误Code
func ErrCode(c int) Option {
	return func(o *Options) {
		o.ErrCode = c
	}
}

// Msg 定义输出信息
func Msg(s string) Option {
	return func(o *Options) {
		o.Msg = s
	}
}

// ErrMsg 定义输出错误信息
func ErrMsg(s string) Option {
	return func(o *Options) {
		o.ErrMsg = s
	}
}

// New 定义新的错误信息
func New(opts ...Option) Service {
	o := &Options{
		ErrCode: ErrCodeDefault,
		ErrMsg:  ErrMsgDefault,
		Msg:     MsgDefault,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Set 设置
func (o *Options) Set(opts ...Option) Service {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Return 返回
func (o Options) Return() error {
	return o
}

// Error error接口实现
func (o Options) Error() string {
	s, _ := json.Marshal(o)
	return string(s)
}

// GetErrCode 获取errCode
func (o Options) GetErrCode() int {
	return o.ErrCode
}

// GetErrMsg 获取errMsg
func (o Options) GetErrMsg() string {
	return o.ErrMsg
}

// GetMsg 获取msg
func (o Options) GetMsg() string {
	return o.Msg
}
