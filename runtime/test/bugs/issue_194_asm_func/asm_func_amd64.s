// +build !noasm,amd64
//go:build !noasm && amd64

#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

TEXT ·AsmFunc2(SB), NOSPLIT, $0-0
    RET
