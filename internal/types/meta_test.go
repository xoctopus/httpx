package types_test

import (
	"fmt"
	"testing"

	. "github.com/xoctopus/httpx/internal/types"
)

func ExampleServerMeta() {
	sm := ServerMeta{
		Name: "server-name",
		// Version: "v0.0.1",
	}
	fmt.Println(sm.UA())
	sm.Version = "v0.0.1"
	fmt.Println(sm.UA())

	rm := RequestMeta{
		Method: "GET",
		Route:  "/path/to/route",
	}
	om := OperationMeta{
		ServerMeta:  sm,
		RequestMeta: rm,
	}
	fmt.Println(om.UA())
	om.ID = "OperatorX"
	fmt.Println(om.UA())

	// Output:
	// server-name
	// server-name@v0.0.1
	// server-name@v0.0.1(unknown)
	// server-name@v0.0.1(OperatorX)
}

func TestExposed(t *testing.T) {
	t.Log(PkgRoot)
	t.Log(ExposedRoot)
}
