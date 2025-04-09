package embed_other_struct

import "github.com/xhd2015/xgo/runtime/test/mock/mock_third_party/third/embed_other_struct/other"

type EmbedStruct struct {
	*other.Other
}
