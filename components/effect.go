package components

// Effect describes an intended state change. Only the runtime applies effects.
type Effect interface {
	isEffect()
}

type effect struct{}

func (effect) isEffect() {}

// SetLocal mutates session-local key-value state.
type SetLocal struct {
	effect
	Key   string
	Value any
}

// Transition moves the session to another scene.
type Transition struct {
	effect
	To SceneID
}

// Persist writes identity-scoped storage.
type Persist struct {
	effect
	Key   string
	Value any
}

// JoinSpace adds the session to a space.
type JoinSpace struct {
	effect
	Space SpaceID
}

// LeaveSpace removes the session from a space.
type LeaveSpace struct {
	effect
	Space SpaceID
}

// SpaceSet sets a key in shared space state (last-write-wins).
type SpaceSet struct {
	effect
	Space SpaceID
	Key   string
	Value any
}

// SpaceCAS sets a key if the current value matches expect (optional strict mode).
type SpaceCAS struct {
	effect
	Space  SpaceID
	Key    string
	Expect any
	Value  any
}

// Broadcast requests subscribed sessions to re-render (space-scoped fan-out).
type Broadcast struct {
	effect
	Space SpaceID
}

// AppendSpaceList appends a value to a list stored under key in space state.
type AppendSpaceList struct {
	effect
	Space SpaceID
	Key   string
	Value any
}

// ReserveNickname records a unique nickname for the current identity.
type ReserveNickname struct {
	effect
	Name string
}

// RegisterAccount creates an account when the runtime account store is enabled.
type RegisterAccount struct {
	effect
	Username string
	Password string
}

// Effect helpers for action/scene authors.

func SetLocalEffect(key string, value any) Effect {
	return SetLocal{Key: key, Value: value}
}

func TransitionEffect(to SceneID) Effect {
	return Transition{To: to}
}

func PersistEffect(key string, value any) Effect {
	return Persist{Key: key, Value: value}
}

func JoinSpaceEffect(space SpaceID) Effect {
	return JoinSpace{Space: space}
}

func LeaveSpaceEffect(space SpaceID) Effect {
	return LeaveSpace{Space: space}
}

func SpaceSetEffect(space SpaceID, key string, value any) Effect {
	return SpaceSet{Space: space, Key: key, Value: value}
}

func SpaceCASEffect(space SpaceID, key string, expect, value any) Effect {
	return SpaceCAS{Space: space, Key: key, Expect: expect, Value: value}
}

func BroadcastEffect(space SpaceID) Effect {
	return Broadcast{Space: space}
}

func AppendSpaceListEffect(space SpaceID, key string, value any) Effect {
	return AppendSpaceList{Space: space, Key: key, Value: value}
}

func ReserveNicknameEffect(name string) Effect {
	return ReserveNickname{Name: name}
}

func RegisterAccountEffect(username, password string) Effect {
	return RegisterAccount{Username: username, Password: password}
}
