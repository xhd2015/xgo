package assert

import "fmt"

func tryDiffError(expected interface{}, actual interface{}) (string, bool) {
	expectedErr, expectedOK := expected.(error)
	actualErr, actualOK := actual.(error)
	if expectedOK != actualOK {
		if expectedOK {
			return fmt.Sprintf("expect: error %s, actual: %T %v", expectedErr.Error(), actual, actual), true
		}
		return fmt.Sprintf("expect: %T %v, actual: error %s", expectedErr, expectedErr, actualErr.Error()), true
	}
	if !expectedOK {
		return "", false
	}
	res := diffError(expectedErr, actualErr)
	if res == nil {
		return "", true
	}
	return res.Error(), true
}

func diffError(expected error, actual error) error {
	if expected == actual {
		return nil
	}
	if expected == nil {
		if actual != nil {
			return fmt.Errorf("expect no error, actual: %v", actual)
		}
		return nil
	}
	if actual == nil {
		return fmt.Errorf("expect error: %v, actual nil", expected)
	}

	e := expected.Error()
	a := actual.Error()
	if e != a {
		return fmt.Errorf("expect err: %v, actual: %v", e, a)
	}
	return nil
}
