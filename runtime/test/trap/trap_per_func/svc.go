// Package trap_per_func checks that registering a mock or interceptor for
// one function lets only that function call the trap: every other
// instrumented function keeps skipping it.
package trap_per_func

import "strconv"

type Client struct{ prefix string }

func (c *Client) Fetch(id int) (string, error) {
	return c.prefix + strconv.Itoa(id), nil
}

func (c *Client) Rank(id int) (int, error) {
	return id * 2, nil
}

func Price(cents int64) (int64, error) {
	return cents * 2, nil
}

func Tax(cents int64) (int64, error) {
	return cents / 10, nil
}

func Untouched(n int) (int, error) {
	return n + 1, nil
}

var Limit = 10

var InitLimit = 10

func ReadLimits() (int, int) {
	return Limit, InitLimit
}
