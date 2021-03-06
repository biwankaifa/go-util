package response

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tracinglog "github.com/opentracing/opentracing-go/log"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	c          *gin.Context
	ErrMsg     string      `json:"err_msg"`
	ErrCode    int         `json:"err_code"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data"`
	StatusCode int
}

// Json json字符串组装
func (r *Response) Json() {

	// 计算最后运行时间
	st, _ := r.c.Get("StateTime")
	if st != nil {
		stateTime := st.(time.Time)
		elapsed := time.Since(stateTime)
		r.c.Writer.Header().Set("X-Runtime", fmt.Sprintln(elapsed))
	}

	r.c.JSON(r.StatusCode, struct {
		ErrMsg  string      `json:"err_msg"`
		ErrCode int         `json:"err_code"`
		Msg     string      `json:"msg"`
		Data    interface{} `json:"data"`
	}{
		ErrMsg:  r.ErrMsg,
		ErrCode: r.ErrCode,
		Msg:     r.Msg,
		Data:    r.Data,
	})
}

// Success 成功返回
func Success(c *gin.Context, data interface{}) {
	r := Response{
		c:          c,
		ErrMsg:     "",
		ErrCode:    0,
		Msg:        "",
		Data:       data,
		StatusCode: http.StatusOK,
	}

	if opentracing.IsGlobalTracerRegistered() {
		span := opentracing.SpanFromContext(c.Request.Context())

		span.LogFields(tracinglog.Object("return", map[string]interface{}{
			"data": data,
		}))
	}

	r.Json()

	log.Printf("Route, path: %s, Data: %v", c.Request.URL, r.Data)
}

// Error 错误返回
func Error(c *gin.Context, err error, parameter ...interface{}) {

	var (
		ErrMsg  = err.Error()
		ErrCode = 400000
		Msg     = "内部错误"
	)

	// 判断error是不是一个json
	if json.Valid([]byte(err.Error())) {
		m, _ := jsonStringToMap(err.Error())
		ErrMsg = mapExist(m, "err_msg")
		ErrCode, _ = strconv.Atoi(mapExist(m, "err_code"))
		Msg = mapExist(m, "msg")
	}

	var data interface{} = nil
	if len(parameter) > 0 && parameter[0] != nil {
		data = parameter[0]
	}

	r := Response{
		c:          c,
		ErrMsg:     ErrMsg,
		ErrCode:    ErrCode,
		Msg:        Msg,
		Data:       data,
		StatusCode: http.StatusOK,
	}

	if opentracing.IsGlobalTracerRegistered() {
		span := opentracing.SpanFromContext(c.Request.Context())

		ext.Error.Set(span, true)
		span.LogFields(tracinglog.Object("return", map[string]interface{}{
			"err_code": ErrCode,
			"err_msg":  ErrMsg,
			"msg":      Msg,
		}))
	}

	r.Json()
}

// mapExist 检查map里面是否存在某个key，返回字符串
func mapExist(m map[string]interface{}, key string) string {
	if _, ok := m[key]; ok {
		return fmt.Sprintf("%v", m[key])
	} else {
		return ""
	}
}

// jsonStringToMap 解析json字符串成 map
func jsonStringToMap(jsonStr string) (m map[string]interface{}, err error) {
	a := map[string]interface{}{}
	unmarsha1Err := json.Unmarshal([]byte(jsonStr), &a)
	if unmarsha1Err != nil {
		return nil, unmarsha1Err
	}
	return a, nil
}
