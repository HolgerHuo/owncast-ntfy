package main

type OwncastPayload struct {
	Type      string       `json:"type"`
	EventData OwncastEvent `json:"EventData"`
}

type OwncastEvent struct {
	Body        string      `json:"body"`
	User        OwncastUser `json:"user"`
	NewName     string      `json:"newName"`
	StreamTitle string      `json:"streamTitle"`
	Summary     string      `json:"summary"`
}

type OwncastUser struct {
	DisplayName   string   `json:"displayName"`
	PreviousNames []string `json:"previousNames"`
}

type NtfyNotification struct {
	Message string   `json:"message"`
	Topic   string   `json:"topic"`
	Title   string   `json:"title"`
	Tags    []string `json:"tags"`
}
