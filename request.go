package elsearm

import (
	"context"

	"github.com/elastic/go-elasticsearch/esapi"
)

// A interface of Request defined by esapi.
type Request interface {
	Do(ctx context.Context, transport esapi.Transport) (*esapi.Response, error)
}

// Zero returns a pointer of int. The value of address is zero.
func Zero() *int {
	var zero int = 0
	return &zero
}
