package patch_arbitrary_stdlib

import (
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
)

// build with --trap-stdlib=false

func TestOsGetEnv(t *testing.T) {
	mock.Patch(os.Getenv, func(key string) string {
		return "MOCK_NOTHING"
	})
	res := os.Getenv("SOMETHING_NEVER_EXISTS")
	expect := "MOCK_NOTHING"
	if res != expect {
		t.Fatalf("expect patched result to be %q, actual: %q", expect, res)
	}
}

func TestPatchArbitraryStdlib(t *testing.T) {
	const src = "hello"
	before := hex.EncodeToString([]byte(src))
	expectBefore := "68656c6c6f"
	if before != expectBefore {
		t.Fatalf("expect before to be %q, actual: %q", expectBefore, before)
	}
	mock.Patch(hex.EncodeToString, func(src []byte) string {
		return "mock " + before
	})
	res := hex.EncodeToString([]byte("hello"))
	expectAfter := "mock 68656c6c6f"
	if res != expectAfter {
		t.Fatalf("expect patched result to be %q, actual: %q", expectAfter, res)
	}
}

func TestIoRead(t *testing.T) {
	mock.Patch(io.ReadAll, func(r io.Reader) ([]byte, error) {
		return []byte("mock"), nil
	})
	res, err := io.ReadAll(strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("expect no error, actual: %v", err)
	}
	expect := "mock"
	if string(res) != expect {
		t.Fatalf("expect patched result to be %q, actual: %q", expect, res)
	}
}

func TestOsExecCombinedOutputFunction(t *testing.T) {
	cmd := exec.Command("ls", "-l")
	mock.Patch((*exec.Cmd).CombinedOutput, func(cmd *exec.Cmd) ([]byte, error) {
		return []byte("mock"), nil
	})
	res, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expect no error, actual: %v", err)
	}
	expect := "mock"
	if string(res) != expect {
		t.Fatalf("expect patched result to be %q, actual: %q", expect, res)
	}
}

func TestOsExecCombinedOutputInstance(t *testing.T) {
	cmd := exec.Command("ls", "-l")
	mock.Patch(cmd.CombinedOutput, func() ([]byte, error) {
		return []byte("mock"), nil
	})
	res, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expect no error, actual: %v", err)
	}
	expect := "mock"
	if string(res) != expect {
		t.Fatalf("expect patched result to be %q, actual: %q", expect, res)
	}
}

func TestNetDial(t *testing.T) {
	myConn := &net.TCPConn{}
	mock.Patch(net.Dial, func(network, addr string) (net.Conn, error) {
		return myConn, nil
	})
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fatalf("expect no error, actual: %v", err)
	}
	if conn != myConn {
		t.Fatalf("expect conn to be %v, actual: %v", myConn, conn)
	}
}

func TestNetHttpClientDo(t *testing.T) {
	client := &http.Client{}
	mock.Patch(client.Do, func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("mock")),
		}, nil
	})
	res, err := client.Do(nil)
	if err != nil {
		t.Fatalf("expect no error, actual: %v", err)
	}
	expect := "mock"
	if res.Body == nil {
		t.Fatalf("expect res.Body to be non-nil,actual nil")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("expect no error, actual: %v", err)
	}
	if string(body) != expect {
		t.Fatalf("expect patched result to be %q, actual: %q", expect, body)
	}
}
