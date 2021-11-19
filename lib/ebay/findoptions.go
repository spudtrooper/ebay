package ebay

// ~/go/bin/genopts -opt_type=FindOption --prefix=Find  page:int seleniumVerbose seleniumHead force

type FindOption func(*findOptionImpl)

type FindOptions interface {
	Page() int
	SeleniumVerbose() bool
	SeleniumHead() bool
	Force() bool
}

func FindPage(page int) FindOption {
	return func(opts *findOptionImpl) {
		opts.page = page
	}
}

func FindSeleniumVerbose(seleniumVerbose bool) FindOption {
	return func(opts *findOptionImpl) {
		opts.seleniumVerbose = seleniumVerbose
	}
}

func FindSeleniumHead(seleniumHead bool) FindOption {
	return func(opts *findOptionImpl) {
		opts.seleniumHead = seleniumHead
	}
}

func FindForce(force bool) FindOption {
	return func(opts *findOptionImpl) {
		opts.force = force
	}
}

type findOptionImpl struct {
	page            int
	seleniumVerbose bool
	seleniumHead    bool
	force           bool
}

func (f *findOptionImpl) Page() int             { return f.page }
func (f *findOptionImpl) SeleniumVerbose() bool { return f.seleniumVerbose }
func (f *findOptionImpl) SeleniumHead() bool    { return f.seleniumHead }
func (f *findOptionImpl) Force() bool           { return f.force }

func makeFindOptionImpl(opts ...FindOption) findOptionImpl {
	var res findOptionImpl
	for _, opt := range opts {
		opt(&res)
	}
	return res
}
