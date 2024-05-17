package logstorage

// filterNoop does nothing
type filterNoop struct {
}

func (fn *filterNoop) String() string {
	return ""
}

func (fn *filterNoop) updateNeededFields(neededFields fieldsSet) {
	// nothing to do
}

func (fn *filterNoop) applyToBlockResult(_ *blockResult, _ *bitmap) {
	// nothing to do
}

func (fn *filterNoop) apply(_ *blockSearch, _ *bitmap) {
	// nothing to do
}
