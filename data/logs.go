package data

// Logs holds all the fields needed for a logs structure
type Logs struct {
	ID      string   `json:"-"`
	Address string   `json:"address"`
	Events  []*Event `json:"events"`
}

// Event holds all the fields needed for an event structure
type Event struct {
	Address    string   `json:"address"`
	Identifier string   `json:"identifier"`
	Topics     [][]byte `json:"topics"`
	Data       []byte   `json:"data"`
}
