package resources

type Group struct {
	Group    string `json:"group"`
	CanWrite bool   `json:"canWrite"`
	CanRead  bool   `json:"canRead"`
}
