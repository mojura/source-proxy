package resources

type Entry struct {
	// Resource name
	Resource string `json:"resource"`
	Groups   Groups `json:"groups"`
}
