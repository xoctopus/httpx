package status_test

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/xoctopus/httpx/internal/status"
	"github.com/xoctopus/httpx/internal/validation"
)

func ExampleAsDescription() {
	// wrapped error
	e := status.WrapCode(errors.New("e1"), http.StatusForbidden)
	de := status.AsDescription(e, "server-name", "query")
	fmt.Println(de)

	// user error
	e = errors.New("e2")
	de = status.AsDescription(e, "server-name", "query")
	fmt.Println(de)

	// validation error
	e = validation.WrapPositionError(validation.WrapLocationError(errors.New("e3"), "query"), "field")
	de = status.AsDescription(e, "server-name", "query")
	fmt.Println(de)
	fmt.Println("  in:  ", de.Location)
	fmt.Println("  pos: ", de.Position)
	fmt.Println("  code:", de.StatusCode())
	fmt.Println("  text:", de.StatusText())

	de2 := status.AsDescription(de, "", "")
	fmt.Println(de2)
	fmt.Println("  in:  ", de2.Location)
	fmt.Println("  pos: ", de2.Position)
	fmt.Println("  code:", de2.StatusCode())
	fmt.Println("  text:", de2.StatusText())

	// Output:
	// FORBIDDEN{message="e1"}
	// INTERNAL_SERVER_ERROR{message="e2"}
	// BAD_REQUEST{message="e3"}
	//   in:   query
	//   pos:  field
	//   code: 400
	//   text: BAD_REQUEST
	// BAD_REQUEST{message="e3"}
	//   in:   query
	//   pos:  field
	//   code: 400
	//   text: BAD_REQUEST
}

func ExampleAsErrorResponse() {
	// e1 := status.WrapStatus(errors.New("e1"), httpx.STATUS__INTERNAL_SERVER_ERROR)
}
