// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

const DefaultDateLayout = "2006-01-02 15:04:05.000Z"

func chainedConst() {
	wrap(DefaultDateLayout).Min(nil)
}

func wrap(e string) interface{ Min(t interface{}) } {
	return nil
}
