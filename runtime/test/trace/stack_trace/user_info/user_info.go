package user_info

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

func GenerateUserName(id int64) string {
	md5 := randMD5()
	return fmt.Sprintf("xhd2015_%d_%s", id, md5)
}

func CheckUserName(name string) error {
	if name == "" {
		return fmt.Errorf("empty user name")
	}
	if !strings.HasPrefix(name, "xhd2015_") {
		return fmt.Errorf("invalid user name")
	}
	return nil
}

func SaveUserName(name string) error {
	if err := CheckUserName(name); err != nil {
		return err
	}
	fmt.Printf("user %s saved\n", name)
	return nil
}

func DeleteUserName(name string) error {
	if name == "xhd2015_1" {
		return fmt.Errorf("cannot delete xhd2015_1")
	}
	fmt.Printf("user %s deleted\n", name)
	return nil
}

func randMD5() string {
	h := md5.New()
	n := generateID()
	h.Write([]byte(strconv.FormatInt(n, 10)))
	return hex.EncodeToString(h.Sum(nil))
}

func generateID() int64 {
	return rand.Int63()
}
