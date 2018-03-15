package xcpb

// ID returns the GPG ID format
func (i Identity) ID() string {
	out := i.Name
	if i.Comment != "" {
		out += " (" + i.Comment + ")"
	}
	out += " <" + i.Email + ">"
	return out
}
