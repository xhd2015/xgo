package init_fn

func init() {

}

func init2() {

}

type SomeType int

func (s SomeType) init() {
	// fine
}

func _() {

}

func __() {

}

func (s SomeType) _() {
	// fine
}
