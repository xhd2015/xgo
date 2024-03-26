package method_recv_via_interface

import (
	"testing"
)

type IService interface {
	GetVersion() string
}

// IService.GetVersion has the following signature
var _ func(IService) string = IService.GetVersion

type service struct {
	version string
}

func (c *service) GetVersion() string {
	return c.version
}

var s = &service{version: "system"}

func GetService() IService {
	return s
}

func TestMethodRecvViaInterface(t *testing.T) {
	s := GetService()

	m := s.GetVersion

	m()
}

// s: rsp+0x10 = typeword, rsp+0x18=dataword
// s is an non-empty interface type
// m: rsp+0
//
// rsp+0x20: fm addr
// rsp+0x28: typeword
// rsp+0x30: dataword
// rdx: ->  rsp+0x20 the context

//  now call the fm
// in X-fm:
//    rsp+0x8: typeword
//    rsp+0x10: dataword
//    X=rsp+0x8 typeword
//    call X(rax=dataword)
//
// X is the underlying struct type, so it will not point to the interface implementation
// actually there is no interface implementation.
// NOTE: a function

// difference:
//    service.M

// IService.M   = {fm addr, real type word, data word}

// an IService to interface{} conversion?
