package snapshot

type _InOut struct {
	Last  string
	Count int
}

type _Response struct {
	Last  string
	Count int
}

func handleL1() *_Response {
	_, resp := handleL2()
	resp.Last = "handleL1"
	resp.Count++
	return resp
}

func handleL2() (*_InOut, *_Response) {
	a := &_InOut{
		Last: "handleL2",
	}
	resp := handleL3(a)

	resp.Last = "handleL2"
	resp.Count++
	return a, resp
}

func handleL3(a *_InOut) *_Response {
	a.Last = "handleL3"
	a.Count++
	return &_Response{
		Last: "handleL3",
	}
}
