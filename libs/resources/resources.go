package resources

import "net/http"

func New(entries ...Entry) *Resources {
	var r Resources
	r.m = make(map[string]Entry, len(entries))
	for _, e := range entries {
		r.m[e.Resource] = e
	}

	return &r
}

type Resources struct {
	m map[string]Entry
}

func (r *Resources) Get(resource string) (e Entry, ok bool) {
	e, ok = r.m[resource]
	return
}

func (r *Resources) GetGroup(resource, group string) (grp Group, ok bool) {
	var e Entry
	if e, ok = r.m[resource]; !ok {
		return
	}

	grp, ok = e.Groups[group]
	return
}

func (g *Resources) Can(httpMethod, resource string, groups ...string) (ok bool) {
	switch httpMethod {
	case http.MethodGet:
		return g.CanRead(resource, groups...)
	case http.MethodPost:
		return g.CanWrite(resource, groups...)
	}

	return false
}

func (g *Resources) CanWrite(resource string, groups ...string) (ok bool) {
	for _, group := range groups {
		var grp Group
		if grp, ok = g.GetGroup(resource, group); !ok {
			return
		}

		if grp.CanWrite {
			return true
		}
	}

	return false
}

func (g *Resources) CanRead(resource string, groups ...string) (ok bool) {
	for _, group := range groups {
		var grp Group
		if grp, ok = g.GetGroup(resource, group); !ok {
			return
		}

		if grp.CanRead {
			return true
		}
	}

	return false
}
