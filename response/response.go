package response

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Response struct {
	ErrMessage string      `json:"errMsg"`
	ErrCode    int         `json:"errCode"`
	Message    string      `json:"msg"`
	Data       interface{} `json:"data"`
}

const (
	Success = 0 // 请求成功

	Err = 40000 // 基础错误

	ErrAuthenticate          = 40005 // 验证器未通过
	ErrValidate              = 40007 // 验证器未通过
	ErrNotFoundRoute         = 40008 // 路由不存在
	ErrNotFoundMedia         = 40002 // 渠道不存在
	ErrDataAbort             = 40003 // 数据异常
	ErrRegisterBindingMobile = 40009 // 此登录方式需绑定手机, 请跳转
	ErrLocationNotFound      = 40010 // 广告位未找到
	ErrContentWordIllegal    = 40011 // 内容存在违规字
	ErrVersionLow            = 40020 // 版本过低, 强制更新
	ErrAccountAbnormality    = 42001 // 账户异常

	ErrPageNotFound = 40004 // authenticate 错误

	ErrNecessaryVip = 41001 // 非会员无权操作
)

var msg = map[int]string{
	Success: "",

	Err: "",

	ErrAuthenticate:       "Authenticate错误",
	ErrValidate:           "数据格式错误",
	ErrNotFoundRoute:      "访问页面不存在",
	ErrNotFoundMedia:      "未查询到该渠道",
	ErrDataAbort:          "内部数据异常",
	ErrLocationNotFound:   "未查询到该广告位数据",
	ErrContentWordIllegal: "内容存在违规字",

	ErrPageNotFound: "Not Found Page",
}

func getMsg(Code int) string {
	return msg[Code]
}

func Json(c *gin.Context, code int, errMsg string, data interface{}, httpStatusCode ...int) {

	var httpCode int

	if len(httpStatusCode) <= 0 {
		httpCode = http.StatusOK
	} else {
		httpCode = httpStatusCode[0]
	}

	st, _ := c.Get("StateTime")
	if st != nil {
		stateTime := st.(time.Time)
		elapsed := time.Since(stateTime)
		c.Writer.Header().Set("X-Runtime", fmt.Sprintln(elapsed))
	}

	c.JSON(httpCode, Response{
		ErrMessage: errMsg,
		ErrCode:    code,
		Message:    getMsg(code),
		Data:       data,
	})

	//log.Printf("Route, path: %s, Data: %v", c.Request.URL, data)
}

func ResponseHTTP(c *gin.Context, data string) {
	c.String(http.StatusOK, data)
}
