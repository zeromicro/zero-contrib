package govalidator

import (
	"context"
	"fmt"
	"testing"
)

type User struct {
	UserId    int64
	PassWord  string `validate:"min=6,max=20"`
	UserNick  string `validate:"min=4,max=15"`
	UserPhone string `validate:"phone"`
	UserEmail string `validate:"email"`
}

func TestValid(t *testing.T) {
	user := User{
		UserId:    1,
		PassWord:  "123",
		UserNick:  "fjwenfjwkfnkwnefjwnfwjkf",
		UserPhone: "12g2f",
		UserEmail: "lijfwf",
	}
	err := DefaultGetValidParams(context.Background(), user)
	if err != nil {
		fmt.Println(err) // PassWord长度必须至少为6个字符,UserNick长度不能超过15个字符,UserPhone格式不正确，必须为手机号码,UserEmail必须是一个有效的邮箱
	}
}
