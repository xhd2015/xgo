package pkg

type unexported struct {
	val int
}

func NewUnexported(n int) *unexported {
	return &unexported{val: n}
}

func (u *unexported) Get() int {
	return u.val
}
