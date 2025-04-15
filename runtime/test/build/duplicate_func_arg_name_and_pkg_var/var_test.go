package duplicate_func_arg_name_and_pkg_var

import "testing"

var tableName = map[string]string{
	"user": "user",
}

func AddCount(dbname, schameName, tableName string, rowsCount int, eventCount bool) {
	var v map[string]string

	// this tableName should not be
	// rewritten
	_ = v[tableName]
}

func TestAddCount(t *testing.T) {
	// should build
	AddCount("dbname", "schameName", "tableName", 1, true)
}
