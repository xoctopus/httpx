package metadata_test

import (
	"fmt"

	. "github.com/xoctopus/httpx/internal/payload/metadata"
)

func ExampleMetadata() {
	m := Merge(Metadata{"k1": []string{"v1"}})
	fmt.Println(m)

	m.Add("k1", "v2")
	fmt.Println(m)

	m.Set("k1", "v3")
	fmt.Println(m)
	m.Add("k2", "v1")
	fmt.Println(m)

	fmt.Println(m.Get("k1"))
	m.Del("k2")
	fmt.Println(m.Get("k2"))
	fmt.Println("K1 exists:", m.Has("k1"))
	fmt.Println("K2 exists:", m.Has("k2"))

	// Output:
	// k1=v1
	// k1=v1&k1=v2
	// k1=v3
	// k1=v3&k2=v1
	// v3
	//
	// K1 exists: true
	// K2 exists: false
}
