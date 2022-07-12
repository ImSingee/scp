package scp

type Options struct {
	scpPath string

	notDefault bool
}

func (o *Options) ScpPath(scpPath string) *Options {
	o.scpPath = scpPath
	return o
}

func (o *Options) applyDefault() *Options {
	if o.notDefault {
		return o
	}

	o.notDefault = true

	if o.scpPath == "" {
		o.scpPath = "scp"
	}

	return o
}

var DefaultOptions = &Options{}
