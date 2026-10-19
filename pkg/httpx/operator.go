package httpx

import (
	"github.com/xoctopus/httpx/internal/operator"
)

type (
	Operator = operator.Operator

	OperatorInitializer   = operator.Initializer
	OperatorDefaultSetter = operator.DefaultSetter
	OperatorContextKey    = operator.HasContextKey
	OperatorMiddlewares   = operator.HasMiddlewares
)
