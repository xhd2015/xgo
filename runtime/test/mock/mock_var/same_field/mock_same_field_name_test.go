package same_field

import (
	"testing"
)

func TestSameFieldNameShouldCompile(t *testing.T) {
	if DB != 10 {
		t.Errorf("expect sub.DB to be 10, but got %d", DB)
	}
	db := Run()
	if db.DB != "test" {
		t.Errorf("expect db.DB to be test, but got %s", db.DB)
	}
}
