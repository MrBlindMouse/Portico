package experience

// Space describes shared multi-user context semantics at definition time.
// Runtime owns membership and broadcasting.
type Space struct {
	OnJoin  []Effect
	OnLeave []Effect
}
