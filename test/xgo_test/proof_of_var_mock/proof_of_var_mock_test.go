// var mock:
//
//		if variable is not a pointer, taking its address, then use it as receiver, it can be converted to value receiver implicitly
//	    otherwise if variable is a pointer, taking its address does not work.
package proof_of_var_mock

import (
	"testing"
)

type Stub struct {
}

type PtrStub *Stub

func (c Stub) A() {
}
func (c *Stub) B() {
}

// invalid receiver type PtrStub (pointer or interface type)
//
// func (c PtrStub) X(){
//
// }

var stub Stub = Stub{}
var pstub *Stub = &Stub{}

func TestVariableCanBeTrapped(t *testing.T) {

	stub.A()
	stub.B()

	__mock_stub := &stub
	__mock_stub.A()
	__mock_stub.B()

	pstub.A()
	pstub.B()

	// __mock_pstub.B undefined (type **Stub has no field or method
	//
	// __mock_pstub := &pstub
	// __mock_pstub.A()
	// __mock_pstub.B()
}
