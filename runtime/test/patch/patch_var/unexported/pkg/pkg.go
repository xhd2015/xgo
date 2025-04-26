package pkg

type xgoTestPatchVarUnexported struct {
	val int
}

func NewUnexported(n int) *xgoTestPatchVarUnexported {
	return &xgoTestPatchVarUnexported{val: n}
}

func (u *xgoTestPatchVarUnexported) Get() int {
	return u.val
}
