// Package scanner helps to scan struct for collection request info
// eg:
//
//	type InBody struct {
//		Name    string `json:"name,nocase,strictcase"`
//		Data    []byte `json:"data"`
//		Ignored any    `name:"-"`
//		NoTag   any
//		Strings []string
//	}
//	type InPath struct {
//		OrgID  int `in:"path" name:"orgID"`
//		UserID int `in:"path" name:"userID"`
//	}
//	type InQuery struct {
//		Q0 int    `in:"query" name:"q0,format:bin"`
//		Q1 string `in:"query" name:"q1"`
//	}
//	type InHeader struct {
//		V1 int    `in:"header" name:"k1"`
//		V2 string `in:"header" name:"k2"`
//	}
//	type InCookie struct {
//		Token string `in:"cookie" name:"token"`
//		CookiePayload
//	}
//	type CookiePayload struct {
//		Userdata string `in:"cookie" name:"userdata"`
//	}
//	type Request struct {
//		httpx.MethodGet `path:"/org/{orgID}/user/{userID}"`
//
//		Direct bool `in:"query" name:"direct"`
//		InCookie
//		InHeader
//		InPath
//		*InQuery
//		InBody  `in:"body" mime:"application/json"`
//	}
//
// Request meta:
// 1. `in` tag describes data location. see: payload.Locations
// 2. `name` tag describes data key.
// 3. `path` tag describes route path.
// 4. json/v2 tag extension is supported. eg: inline, unknown, string, format etc.
//
// eg: request content
// GET /org/1001/user/8888?direct=true&q0=1010&q1=active HTTP/1.1
// Host: api.yourgateway.com
// k1: 200
// k2: high-priority
// Cookie: token=abc_session_99; userdata=gold_member
// Content-Type: application/json
//
//	{
//	    "name": "xoctopus",
//	    "data": "SGVsbG8gV29ybGQ=",
//	    "strings": ["opt1", "opt2"]
//	}
//
// Collect results:
// 1. InPath{ OrgID: 1; UserID: 2}
// 2. InQuery{ Q0: 101; Q1: "active"}; Direct: true
// 3. InHeader{ V1: 200; V2: "high-priority" }
// 4. InCookie{ Token: "abc_session_99" Userdata: "gold_member" }
// 5. InBody{ Name: "xoctopus"; Data: []byte("Hello World"); Strings: []string{"opt1","opt2"}
package scanner
