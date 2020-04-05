package plex

type ServerMetric struct {
	Version        string
	Platform       string
	ActiveSessions int
	Players        []PlayerMetric
	Libraries      []LibraryMetric
}

type LibraryMetric struct {
	Name string
	Type string
	Size int
}

type PlayerMetric struct {
	Device   string
	Platform string
	Profile  string
	State    string
	Local    string
	Relayed  string
	Secure   string
}
