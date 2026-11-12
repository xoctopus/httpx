// Package httpx exports the building blocks of httpx services: operators,
// routers, requests and responses.
//
// # Output
//
// An operator returns (any, error) from Output. The response of each form:
//
//   - nil, nil: 204 without body.
//   - data, nil: 201 for POST, otherwise 200; the body is data serialized by its
//     type (JSON for structs).
//   - WrapResponse(v, ...), nil: like data, nil; the options set the status code,
//     headers, cookies and Content-Type.
//   - nil, codex error: an RFC 9457 problem (application/problem+json). the
//     status is the first one declared from outside in (`@code api`, [Status]
//     Wrap); code, key, title and detail come from the outermost codex error.
//   - nil, plain error: a problem with status 500, or the status declared by
//     [Status].Wrap; the error text is never exposed.
//   - ErrorResponse(err, ...), nil: the problem of err, with the options applied.
//
// An error, or an [ErrorResponse] returned as value, interrupts the operator
// chain: following operators are not executed. This holds for pre operators.
//
// # Problem Details
//
// Error responses carry `code`, `key`, `title`, `detail` and `errors`; empty
// fields are omitted, `type`, `status` and `instance` are never written. The full
// text of errors (codex messages, causes) only goes to logs.
//
// # Response Options
//
// [WithResponseStatusCode], [WithResponseMetadata], [WithResponseCookies] and
// [WithResponseContentType] apply to both [WrapResponse] and [ErrorResponse].
// [WithResponseTitle] and [WithResponseDetail] override the problem fields of an
// [ErrorResponse], for errors not built with codex.
package httpx
