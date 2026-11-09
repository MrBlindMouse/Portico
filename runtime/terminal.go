package runtime

// Terminal abstracts SSH or fake terminal I/O for the session loop.
type Terminal interface {
	WriteFrame(text string) error
	ReadInput() (string, error)
	OnResize(func(width, height int))
	Close() error
}
