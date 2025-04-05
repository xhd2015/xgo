package recover_no_trap

func A() {
	defer B()
}

func B() string {
	recover()
	return "B"
}
