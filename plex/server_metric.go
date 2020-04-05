package plex

type ServerMetric struct {
	Version        string
	Platform       string
	ActiveSessions int
	Sessions       []SessionMetric
	Players        []PlayerMetric
	Libraries      []LibraryMetric
}

type LibraryMetric struct {
	Name string
	Type string
	Size int
}

type SessionMetric struct {
	Location string
}

type PlayerMetric struct {
	Device   string
	Platform string
	Profile  string
	State    string
	Local    bool
	Relayed  bool
	Secure   bool
}
