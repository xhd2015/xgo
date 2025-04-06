package id

var id int

func Next() int {
	id++
	return id
}
