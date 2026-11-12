// Package status resolves errors into RFC 9457 style problem details.
//
// [Response] is the root written to the client; it is the exposed subset of
// [Detail], the full problem that travels as an error. They stay separate types
// so that the exposed surface can hide more in the future.
//
// # Resolution
//
// [AsResponse] and [AsDetail] walk an error chain from outside in, and resolve
// two things independently, each from the first node providing it:
//
//   - status: the first node declaring a status code (> 0), such as
//     [Wrap]/[WrapCode] or a codex code declaring `@code api`. nodes declaring
//     nothing are skipped. without any declaration it is 400 for errors located
//     in the request ([HasPointer]), otherwise 500.
//   - identity layer: the outermost codex.Error. code, key, title and detail
//     come from it only: code is codex.Error.Code, key is the code identifier,
//     title is the code hint and detail is codex.Error.Detail (the format and
//     args of codex.Errorf / codex.Wrapf). fields it lacks are left empty; they
//     are neither looked up from inner nodes nor filled with the status text.
//
// Each branch of errors.Join becomes an entry of Errors: branches inherit the
// status above the Join and resolve their own identity layer. The top level of
// [Response] takes the identity layer above the Join if any, otherwise the
// first entry. A downstream [*Detail] in the chain is kept as an entry as is.
//
// # Exposure
//
// The full text of an error (codex messages, causes, Error() of plain errors)
// is kept in Desc, which is never serialized (`json:"-"`) and only serves
// logging. Response bodies carry no `type`, `status` or `instance`: the status
// is in the status line, the identity is `key` and the source is the `Server`
// response header.
package status
