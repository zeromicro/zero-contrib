package gin

import "errors"

var (
	// ErrInvalidMethod is an error that indicates not a valid http method.
	ErrInvalidMethod = errors.New("not a valid http method")
	// ErrInvalidPath is an error that indicates path is not start with /.
	ErrInvalidPath = errors.New("path must begin with '/'")
)

type config struct {
	redirectTrailingSlash bool
	redirectFixedPath     bool
}

func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

type Option func(o *config)

func WithRedirectTrailingSlash(redirectTrailingSlash bool) Option {
	return func(c *config) {
		c.redirectTrailingSlash = redirectTrailingSlash
	}
}

func WithRedirectFixedPath(redirectFixedPath bool) Option {
	return func(c *config) {
		c.redirectFixedPath = redirectFixedPath
	}
}
