package resources

import "encoding/json"

type Groups map[string]Group

func (g Groups) MarshalJSON() (bs []byte, err error) {
	gs := make([]Group, 0, len(g))
	for _, grp := range g {
		gs = append(gs, grp)
	}

	return json.Marshal(gs)
}

func (g *Groups) UnmarshalJSON(bs []byte) (err error) {
	var gs []Group
	if err = json.Unmarshal(bs, &gs); err != nil {
		return
	}

	m := make(Groups, len(gs))
	for _, grp := range gs {
		m[grp.Group] = grp
	}

	*g = m
	return
}
