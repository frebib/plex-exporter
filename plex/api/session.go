package api

type SessionList struct {
	Sessions `json:"MediaContainer"`
}

type Sessions struct {
	Size     int               `json:"size"`
	Metadata []SessionMetadata `json:"Metadata"`
}

type SessionMetadata struct {
	Session `json:"Session"`
	Player  `json:"Player"`
}

type Session struct {
	Bandwidth int    `json:"bandwidth"`
	Location  string `json:"location"`
}

type Player struct {
	Device   string `json:"device"`
	Platform string `json:"platform"`
	Profile  string `json:"profile"`
	State    string `json:"state"`
	Local    bool   `json:"local"`
	Relayed  bool   `json:"relayed"`
	Secure   bool   `json:"secure"`
}
