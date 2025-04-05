package stack_trace

import (
	"fmt"

	"github.com/xhd2015/xgo/runtime/test/trace/stack_trace/user_info"
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

func DeleteUserInfo(name string) {
	defer func() {
		if e := recover(); e != nil {
			panic(fmt.Errorf("something bad: %v", e))
		}
	}()
	doDeleteUserInfo(name)
}
func doDeleteUserInfo(name string) {
	checkErr := user_info.CheckUserName(name)
	if checkErr != nil {
		panic(checkErr)
	}

	err := user_info.DeleteUserName(name)
	if err != nil {
		panic(err)
	}
}
