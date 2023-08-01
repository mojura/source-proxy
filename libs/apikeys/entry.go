package apikeys

type Entry struct {
	// Name of the Entry, this is optional and is just used to make entries
	// easier to differ between.
	Name string `json:"name"`
	// APIKey of entry
	APIKey string `json:"apikey"`
	// Groups entry is a part of
	Groups []string `json:"groups"`
}
