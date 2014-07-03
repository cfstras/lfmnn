package load

type Loader interface {
}

type loader struct {
	apikey, secret string
}

func NewLoader(apikey, secret string) Loader {
	l := &loader{apikey: apikey, secret: secret}
	return l
}
