package internal

type Round struct {
	Phase   string `json:"phase,omitempty"`
	WinTeam string `json:"win_team,omitempty"`
}

type Previously struct {
	Round Round `json:"round,omitempty"`
}

type GameState struct {
	Round      Round      `json:"round"`
	Previously Previously `json:"previously"`
}
