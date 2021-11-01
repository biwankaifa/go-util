package validator

type Exception struct {
	Msg     string
	ErrMsg  string
	ErrCode int
}

var ErrValidate = 400000
