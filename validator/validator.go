package validator

import (
	"context"
	"errors"
	"github.com/biwankaifa/go-util/response"
	enLanguage "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	"reflect"
)

var (
	uni      *ut.UniversalTranslator
	validate *validator.Validate
	trans    ut.Translator
)

func init() {
	validate = validator.New()
	en := enLanguage.New()
	uni = ut.New(en)
	trans, _ := uni.GetTranslator("en")
	validate = validator.New()
	//注册一个函数，获取struct tag里自定义的label作为字段名
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		return name
	})
	//注册翻译器
	_ = enTranslations.RegisterDefaultTranslations(validate, trans)
}

func VarPanic(v interface{}, tag string) {
	if err := Var(v, tag); err != nil {
		panic(Exception{
			Msg:  err.Error(),
			Code: response.ErrValidate,
		})
	}
}

func StructPartialPanic(v interface{}, fields ...string) {
	if err := StructPartial(v, fields...); err != nil {
		panic(Exception{
			Msg:  err.Error(),
			Code: response.ErrValidate,
		})
	}
}

func StructPanic(v interface{}) {
	if err := Struct(v); err != nil {
		panic(Exception{
			Msg:  err.Error(),
			Code: response.ErrValidate,
		})
	}
}

func Var(field interface{}, tag string) error {
	if err := validate.VarCtx(context.Background(), field, tag); err != nil {
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			return errors.New(e.Translate(trans))
		}
	}
	return nil
}

func StructPartial(s interface{}, fields ...string) error {
	if err := validate.StructPartialCtx(context.Background(), s, fields...); err != nil {
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			return errors.New(e.Translate(trans))
		}
	}
	return nil
}

func Struct(v interface{}) error {
	if err := validate.Struct(v); err != nil {
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			return errors.New(e.Translate(trans))
		}
	}
	return nil
}
