package govalidator

import (
	"context"
	"errors"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"reflect"
	"regexp"
	"strings"
)

const (
	ValidatorKey  = "ValidatorKey"
	TranslatorKey = "TranslatorKey"
	locale        = "chinese"
)

func TransInit(ctx context.Context) context.Context {
	// Set the supported language
	chinese := zh.New()
	english := en.New()
	// Set up an internationalized translator
	uni := ut.New(chinese, chinese, english)
	// Set up validators
	val := validator.New()
	// Take the translator instance as an argument
	trans, _ := uni.GetTranslator(locale)
	// Register the translator with validatorfw
	switch locale {
	case "chinese":
		zhTranslations.RegisterDefaultTranslations(val, trans)
		// Register a custom method to get a tag with fld.Tag.Get("comment")
		val.RegisterTagNameFunc(func(fld reflect.StructField) string {
			return fld.Tag.Get("comment")
		})
		val.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
			phone := fl.Field().String()
			// Validate phone number with a regular expression
			pattern := `^1[3456789]\d{9}$`
			matched, _ := regexp.MatchString(pattern, phone)
			return matched
		})
		zhTranslations.RegisterDefaultTranslations(val, trans)
		val.RegisterTranslation("phone", trans, func(ut ut.Translator) error {
			return ut.Add("phone", "{0}格式不正确，必须为手机号码", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("phone", fe.Field())
			return t
		})

	case "english":
		enTranslations.RegisterDefaultTranslations(val, trans)
		val.RegisterTagNameFunc(func(fld reflect.StructField) string {
			return fld.Tag.Get("en_comment")
		})
		val.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
			phone := fl.Field().String()
			// Validate phone number with a regular expression
			pattern := `^1[3456789]\d{9}$`
			matched, _ := regexp.MatchString(pattern, phone)
			return matched
		})
		zhTranslations.RegisterDefaultTranslations(val, trans)
		val.RegisterTranslation("phone", trans, func(ut ut.Translator) error {
			return ut.Add("phone", "The {0} format is incorrect;it must be a cell phone number", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("phone", fe.Field())
			return t
		})
	}

	ctx = context.WithValue(ctx, ValidatorKey, val)
	ctx = context.WithValue(ctx, TranslatorKey, trans)
	return ctx
}

func DefaultGetValidParams(ctx context.Context, params interface{}) error {
	ctx = TransInit(ctx)
	err := validate(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func validate(ctx context.Context, params interface{}) error {
	// Get the validatorfw
	val, ok := ctx.Value(ValidatorKey).(*validator.Validate)
	if !ok {
		return errors.New("Validator not found in context")
	}

	// Get the translator
	tran, ok := ctx.Value(TranslatorKey).(ut.Translator)
	if !ok {
		return errors.New("Translator not found in context")
	}
	err := val.Struct(params)
	// If the data validation fails, output all err in a slice
	if err != nil {
		errs := err.(validator.ValidationErrors)
		sliceErrs := []string{}
		for _, e := range errs {
			// use the validatorfw. ValidationErrors types in the `Translate` method for translation
			sliceErrs = append(sliceErrs, e.Translate(tran))
		}
		return errors.New(strings.Join(sliceErrs, ","))
	}
	return nil
}
