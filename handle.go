package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"reflect"
	"time"
)

func GenerateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max-min) + min
	return randNum
}

func Md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func GetOffset(page int, listRows int) int {
	if page == 0 {
		page = 1
	}
	switch {
	case listRows > 100:
		listRows = 100
	case listRows <= 0:
		listRows = 10
	}
	return (page - 1) * listRows

}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// FormatMobileStar 手机号中间4位替换为*号
func FormatMobileStar(mobile string) string {
	l := len(mobile)
	switch {
	case l <= 4:
		return mobile
	case l == 5:
		return mobile[:1] + "****"
	case l == 6:
		return mobile[:1] + "****" + mobile[5:]
	case l == 7 || l == 8:
		return mobile[:2] + "****" + mobile[6:]
	case l >= 9:
		return mobile[:3] + "****" + mobile[7:]
	}
	return mobile
}

// ToInterfaceSlice interface{}转化为[]interface{}
func ToInterfaceSlice(arr interface{}) []interface{} {
	ret := make([]interface{}, 0)
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		ret = append(ret, arr)
		return ret
	}
	l := v.Len()
	for i := 0; i < l; i++ {
		ret = append(ret, v.Index(i).Interface())
	}
	return ret
}

func MapToJson(param map[string]interface{}) string {
	dataType, _ := json.Marshal(param)
	dataString := string(dataType)
	return dataString
}
