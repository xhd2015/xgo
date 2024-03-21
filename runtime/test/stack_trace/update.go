package stack_trace

import (
	"github.com/xhd2015/xgo/runtime/test/stack_trace/user_info"
)

func UpdateUserInfo(name string) (actualName string, err error) {
	actualName = name
	checkErr := user_info.CheckUserName(name)
	if checkErr != nil {
		actualName = user_info.GenerateUserName(1)
	}

	err = user_info.SaveUserName(actualName)
	if err != nil {
		return "", err
	}
	return actualName, nil
}
