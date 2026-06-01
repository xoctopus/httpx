package httpx

import (
	"net/http"
	"strconv"

	_ "github.com/xoctopus/x/enumx"

	"github.com/xoctopus/httpx/internal/status"
)

// Status presents http status code. imported from net/http/status.go
// HTTP status codes as registered with IANA.
// See: https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
// +genx:enum
type Status int

const (
	STATUS_UNKNOWN Status = 0

	STATUS__CONTINUE            Status = http.StatusContinue           // RFC 9110, 15.2.1
	STATUS__SWITCHING_PROTOCOLS Status = http.StatusSwitchingProtocols // RFC 9110, 15.2.2
	STATUS__PROCESSING          Status = http.StatusProcessing         // RFC 2518, 10.1
	STATUS__EARLY_HINTS         Status = http.StatusEarlyHints         // RFC 8297

	STATUS__OK                     Status = http.StatusOK                   // RFC 9110, 15.3.1
	STATUS__CREATED                Status = http.StatusCreated              // RFC 9110, 15.3.2
	STATUS__ACCEPTED               Status = http.StatusAccepted             // RFC 9110, 15.3.3
	STATUS__NON_AUTHORITATIVE_INFO Status = http.StatusNonAuthoritativeInfo // RFC 9110, 15.3.4
	STATUS__NO_CONTENT             Status = http.StatusNoContent            // RFC 9110, 15.3.5
	STATUS__RESET_CONTENT          Status = http.StatusResetContent         // RFC 9110, 15.3.6
	STATUS__PARTIAL_CONTENT        Status = http.StatusPartialContent       // RFC 9110, 15.3.7
	STATUS__MULTI_STATUS           Status = http.StatusMultiStatus          // RFC 4918, 11.1
	STATUS__ALREADY_REPORTED       Status = http.StatusAlreadyReported      // RFC 5842, 7.1
	STATUS__IM_USED                Status = http.StatusIMUsed               // RFC 3229, 10.4.1

	STATUS__MULTIPLE_CHOICES   Status = http.StatusMultipleChoices   // RFC 9110, 15.4.1
	STATUS__MOVED_PERMANENTLY  Status = http.StatusMovedPermanently  // RFC 9110, 15.4.2
	STATUS__FOUND              Status = http.StatusFound             // RFC 9110, 15.4.3
	STATUS__SEE_OTHER          Status = http.StatusSeeOther          // RFC 9110, 15.4.4
	STATUS__NOT_MODIFIED       Status = http.StatusNotModified       // RFC 9110, 15.4.5
	STATUS__USE_PROXY          Status = http.StatusUseProxy          // RFC 9110, 15.4.6
	_                          Status = http.StatusUseProxy + 1      // RFC 9110, 15.4.7 (Unused)
	STATUS__TEMPORARY_REDIRECT Status = http.StatusTemporaryRedirect // RFC 9110, 15.4.8
	STATUS__PERMANENT_REDIRECT Status = http.StatusPermanentRedirect // RFC 9110, 15.4.9

	STATUS__BAD_REQUEST                     Status = http.StatusBadRequest                   // RFC 9110, 15.5.1
	STATUS__UNAUTHORIZED                    Status = http.StatusUnauthorized                 // RFC 9110, 15.5.2
	STATUS__PAYMENT_REQUIRED                Status = http.StatusPaymentRequired              // RFC 9110, 15.5.3
	STATUS__FORBIDDEN                       Status = http.StatusForbidden                    // RFC 9110, 15.5.4
	STATUS__NOTFOUND                        Status = http.StatusNotFound                     // RFC 9110, 15.5.5
	STATUS__METHOD_NOT_ALLOWED              Status = http.StatusMethodNotAllowed             // RFC 9110, 15.5.6
	STATUS__NOT_ACCEPTABLE                  Status = http.StatusNotAcceptable                // RFC 9110, 15.5.7
	STATUS__PROXY_AUTH_REQUIRED             Status = http.StatusProxyAuthRequired            // RFC 9110, 15.5.8
	STATUS__REQUEST_TIMEOUT                 Status = http.StatusRequestTimeout               // RFC 9110, 15.5.9
	STATUS__CONFLICT                        Status = http.StatusConflict                     // RFC 9110, 15.5.10
	STATUS__GONE                            Status = http.StatusGone                         // RFC 9110, 15.5.11
	STATUS__LENGTH_REQUIRED                 Status = http.StatusLengthRequired               // RFC 9110, 15.5.12
	STATUS__PRECONDITION_FAILED             Status = http.StatusPreconditionFailed           // RFC 9110, 15.5.13
	STATUS__REQUEST_ENTITY_TOO_LARGE        Status = http.StatusRequestEntityTooLarge        // RFC 9110, 15.5.14
	STATUS__REQUEST_URI_TOO_LONG            Status = http.StatusRequestURITooLong            // RFC 9110, 15.5.15
	STATUS__UNSUPPORTED_MEDIA_TYPE          Status = http.StatusUnsupportedMediaType         // RFC 9110, 15.5.16
	STATUS__REQUESTED_RANGE_NOT_SATISFIABLE Status = http.StatusRequestedRangeNotSatisfiable // RFC 9110, 15.5.17
	STATUS__EXPECTATION_FAILED              Status = http.StatusExpectationFailed            // RFC 9110, 15.5.18
	STATUS__TEAPOT                          Status = http.StatusTeapot                       // RFC 9110, 15.5.19 (Unused)
	STATUS__MISDIRECTED_REQUEST             Status = http.StatusMisdirectedRequest           // RFC 9110, 15.5.20
	STATUS__UNPROCESSABLE_ENTITY            Status = http.StatusUnprocessableEntity          // RFC 9110, 15.5.21
	STATUS__LOCKED                          Status = http.StatusLocked                       // RFC 4918, 11.3
	STATUS__FAILED_DEPENDENCY               Status = http.StatusFailedDependency             // RFC 4918, 11.4
	STATUS__TOO_EARLY                       Status = http.StatusTooEarly                     // RFC 8470, 5.2.
	STATUS__UPGRADE_REQUIRED                Status = http.StatusUpgradeRequired              // RFC 9110, 15.5.22
	STATUS__PRECONDITION_REQUIRED           Status = http.StatusPreconditionRequired         // RFC 6585, 3
	STATUS__TOO_MANY_REQUESTS               Status = http.StatusTooManyRequests              // RFC 6585, 4
	STATUS__REQUEST_HEADER_FIELDS_TOO_LARGE Status = http.StatusRequestHeaderFieldsTooLarge  // RFC 6585, 5
	STATUS__UNAVAILABLE_FOR_LEGAL_REASONS   Status = http.StatusUnavailableForLegalReasons   // RFC 7725, 3

	STATUS__INTERNAL_SERVER_ERROR           Status = http.StatusInternalServerError           // RFC 9110, 15.6.1
	STATUS__NOT_IMPLEMENTED                 Status = http.StatusNotImplemented                // RFC 9110, 15.6.2
	STATUS__BAD_GATEWAY                     Status = http.StatusBadGateway                    // RFC 9110, 15.6.3
	STATUS__SERVICE_UNAVAILABLE             Status = http.StatusServiceUnavailable            // RFC 9110, 15.6.4
	STATUS__GATEWAY_TIMEOUT                 Status = http.StatusGatewayTimeout                // RFC 9110, 15.6.5
	STATUS__HTTP_VERSION_NOT_SUPPORTED      Status = http.StatusHTTPVersionNotSupported       // RFC 9110, 15.6.6
	STATUS__VARIANT_ALSO_NEGOTIATES         Status = http.StatusVariantAlsoNegotiates         // RFC 2295, 8.1
	STATUS__INSUFFICIENT_STORAGE            Status = http.StatusInsufficientStorage           // RFC 4918, 11.5
	STATUS__LOOP_DETECTED                   Status = http.StatusLoopDetected                  // RFC 5842, 7.2
	STATUS__NOT_EXTENDED                    Status = http.StatusNotExtended                   // RFC 2774, 7
	STATUS__NETWORK_AUTHENTICATION_REQUIRED Status = http.StatusNetworkAuthenticationRequired // RFC 6585, 6

	// Nginx
	// ref: https://en.wikipedia.org/wiki/List_of_HTTP_status_codes#nginx

	STATUS__NO_RESPONSE                     Status = 444
	STATUS__REQUEST_HEADER_TOO_LARGE        Status = 494
	STATUS__SSL_CERTIFICATE_ERROR           Status = 495
	STATUS__SSL_CERTIFICATE_REQUIRED        Status = 496
	STATUS__HTTP_REQUEST_SENT_TO_HTTPS_PORT Status = 497
	STATUS__CLIENT_CLOSED_REQUEST           Status = 499

	// Cloudflare TODO
	// ref: https://en.wikipedia.org/wiki/List_of_HTTP_status_codes#Cloudflare
)

func (s Status) StatusCode() int {
	return int(s)
}

func (s Status) StatusText() string {
	return s.String()
}

func (s Status) IsValid() bool {
	return s > 0
}

func (s Status) Wrap(err error) error {
	if err == nil || IsStatusOK(s) {
		return nil
	}
	return status.Wrap(err, s)
}

func IsStatusOK[Code ~int](v Code) bool {
	i := int(v)
	if i > 999 {
		i, _ = strconv.Atoi(strconv.Itoa(i)[0:3])
	}
	return i >= http.StatusOK && i < http.StatusMultipleChoices
}
