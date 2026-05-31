package testutil_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/xoctopus/httpx/internal/testutil"
)

func ExampleUnifyRequest() {
	body := `{"k1":100,"k2":"200"}`

	req, _ := http.NewRequest(
		"POST",
		"/users/1?k1=v1&k2=v2",
		bytes.NewBufferString(body),
	)
	req.Header.Set("content-type", "applicaiton/json")
	req.Header.Set("content-length", strconv.Itoa(len(body)))
	req.Header.Set("user-agent", "goland")
	req.URL.RawQuery = (url.Values{
		"k1": []string{"1"},
		"k2": []string{"1", "2", "3"},
	}).Encode()

	fmt.Println(testutil.UnifyRequest(req))

	// Output:
	// POST /users/1?k1=1&k2=1&k2=2&k2=3 HTTP/1.1
	// Content-Length: 21
	// Content-Type: applicaiton/json
	// User-Agent: goland
	//
	// {"k1":100,"k2":"200"}
}
