package embed_other_struct

import "github.com/xhd2015/xgo/runtime/test/patch/patch_third_party_var/third/embed_other_struct/other"

type EmbedStruct struct {
	*other.Other
}
