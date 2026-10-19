// Package operator defines the handler pipeline step model for httpx.
//
// A route is a sequence of operators. [route.Handler] runs each step: unmarshal
// the request into a per-request instance, optional [Initializer.Init], then
// [Operator.Output]. Non-final steps store [Output] in context for later steps;
// the final step's result is written as the HTTP response. Registration-time
// metadata and instance creation live in [Factory]. The public alias is
// [httpx.Operator].
package operator
