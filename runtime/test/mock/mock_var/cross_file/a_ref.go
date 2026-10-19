// Package cross_file references variables declared in a later file whose
// types come from imports this file does not have.
package cross_file

func TimeoutSeconds() float64 {
	return timeout.Seconds()
}

func Incr() int {
	mu.Lock()
	defer mu.Unlock()
	count++
	return count
}
