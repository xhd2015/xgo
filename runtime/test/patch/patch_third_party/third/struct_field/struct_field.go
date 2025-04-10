package struct_field

import "github.com/xhd2015/xgo/runtime/test/patch/patch_third_party/third/struct_field/other"

type SomeStruct struct {
	MyField *other.Other
}
