package mousegesturepresentation

// Snapshot is the complete, combined mouse-gesture mutation contract.
type Snapshot struct {
	Enabled bool
	Tilts   []Tilt
}

type Tilt struct {
	ID, Name string
	Value    uint8
}
