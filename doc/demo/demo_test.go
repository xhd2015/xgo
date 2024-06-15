package demo

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

func MyFunc() string {
	return "my func"
}
func TestFuncMock(t *testing.T) {
	mock.Patch(MyFunc, func() string {
		return "mock func"
	})
	text := MyFunc()
	if text != "mock func" {
		t.Fatalf("expect MyFunc() to be 'mock func', actual: %s", text)
	}
}

type Student struct {
	Name string
	Age  int
}

func NewStudent() *Student {
	return &Student{
		Name: "xiaoming",
		Age:  18,
	}
}

func (s *Student) GetName() string {
	return s.Name
}

func (s *Student) SetName(name string) {
	s.Name = name
}

func (s *Student) getAge() int {
	return s.Age
}

func TestStructPatch(t *testing.T) {
	student := NewStudent()
	mock.Patch(student.GetName, func() string {
		return "zhangsan"
	})

	mock.Patch(student.getAge, func() int {
		return 20
	})

	name := student.GetName()
	if name != "zhangsan" {
		t.Fatalf("expect GetName() to be 'zhangsan', actual: %s", name)
	}

	age := student.getAge()
	if age != 20 {
		t.Fatalf("expect getAge() to be '20', actual: %v", age)
	}

	mock.Patch(student.SetName, func(name string) {
		return
	})
	student.SetName("lisi")

	if student.Name != "xiaoming" {
		t.Fatalf("expect Name to be 'xiaoming', actual: %s", student.Name)
	}
}

type IStudent interface {
	GetName() string
}

func TestInterfaceStructPatch(t *testing.T) {
	var iStudent IStudent
	iStudent = NewStudent()
	mock.Patch(iStudent.GetName, func() string {
		return "zhangsan"
	})
	name := iStudent.GetName()
	if name != "zhangsan" {
		t.Fatalf("expect GetName() to be 'zhangsan', actual: %s", name)
	}
}
