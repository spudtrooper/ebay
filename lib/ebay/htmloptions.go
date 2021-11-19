package ebay

// ~/go/bin/genopts -opt_type=HTMLOption --prefix=HTML inlineAssets

type HTMLOption func(*hTMLOptionImpl)

type HTMLOptions interface {
	InlineAssets() bool
}

func HTMLInlineAssets(inlineAssets bool) HTMLOption {
	return func(opts *hTMLOptionImpl) {
		opts.inlineAssets = inlineAssets
	}
}

type hTMLOptionImpl struct {
	inlineAssets bool
}

func (h *hTMLOptionImpl) InlineAssets() bool { return h.inlineAssets }

func makeHTMLOptionImpl(opts ...HTMLOption) hTMLOptionImpl {
	var res hTMLOptionImpl
	for _, opt := range opts {
		opt(&res)
	}
	return res
}
