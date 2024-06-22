package protocol

type options struct {
	version     uint8
	subversion  uint8
	compression bool
}

type Option func(*options)

func Version(version uint8) Option {
	return func(o *options) {
		o.version = version
	}
}

func Subversion(subversion uint8) Option {
	return func(o *options) {
		o.subversion = subversion
	}
}

func Compression(compression bool) Option {
	return func(o *options) {
		o.compression = compression
	}
}

func Options(opts ...Option) *options {
	o := &options{
		version:     1,
		subversion:  0,
		compression: false,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
