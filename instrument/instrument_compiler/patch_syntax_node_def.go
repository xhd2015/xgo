package instrument_compiler

// only missing in go1.21 and below
const NodesGen = `
func (n *node) SetPos(p Pos) {
	n.pos = p
}
`

const Nodes_Inspect_117 = `
// Walk stops when f returns true, so invert it here
func Inspect(root Node, f func(Node) bool) {
	Walk(root, func(n Node) bool {
		return !f(n)
	})
}
`
