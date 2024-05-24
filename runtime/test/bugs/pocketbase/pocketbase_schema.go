// see https://github.com/xhd2015/xgo/issues/142
package pocketbase

const DefaultDateLayout = "2006-01-02 15:04:05.000Z"

// from https://github.com/pocketbase/pocketbase/blob/4937acb3e2685273998506b715e2b54e33174172/models/schema/schema_field.go#L591
func chainedConst() {
	// the problem only appears when chained const with
	// following call with arg
	wrap(DefaultDateLayout).Min(nil)
}

func wrap(e string) interface{ Min(t interface{}) } {
	return nil
}
