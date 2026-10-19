// Package shadowed_init refers to a variable from a function in which a
// local shadows a name the initializer of the variable uses.
package shadowed_init

import "time"

var timeout = newTimeout()

func newTimeout() time.Duration {
	return 2 * time.Second
}

func TimeoutSeconds() float64 {
	newTimeout := 0
	_ = newTimeout
	return timeout.Seconds()
}
