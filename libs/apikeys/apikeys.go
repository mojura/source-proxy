package apikeys

func New(entries ...Entry) *APIKeys {
	var a APIKeys
	a.m = make(map[string]Entry, len(entries))
	for _, e := range entries {
		a.m[e.APIKey] = e
	}

	return &a
}

type APIKeys struct {
	m map[string]Entry
}

func (a *APIKeys) Get(apikey string) (e Entry, ok bool) {
	e, ok = a.m[apikey]
	return
}

func (a *APIKeys) Groups(apikey string) (groups []string) {
	e, ok := a.Get(apikey)
	if !ok {
		return
	}

	groups = e.Groups
	return
}
