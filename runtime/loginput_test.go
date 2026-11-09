package runtime

import (
	"strings"
	"testing"
)

func TestLogInputPreview(t *testing.T) {
	short := "guestbook"
	if logInputPreview(short) != short {
		t.Fatal(short)
	}
	long := strings.Repeat("a", 80)
	p := logInputPreview(long)
	if len([]rune(p)) > logInputPreviewRunes+1 {
		t.Fatalf("preview too long: %d", len([]rune(p)))
	}
}

func TestProtocolLineKind(t *testing.T) {
	if protocolLineKind("KITTY_DATA_START") != "kitty" {
		t.Fatal("kitty prefix")
	}
	if protocolLineKind("OK") != "protocol" {
		t.Fatalf("short ack kind=%q", protocolLineKind("OK"))
	}
	b64 := "n0v4uazhRRJ+tvBzET+X8HNZw3sg/Gzh5yJ+LuHnsnb8fAf+L8BHCz4W4WMJPpY1EIlqKAvVQFSoCbGb+O5rIFd6J4l59/fu793fu78f9u//HwAzKeTyAH4BAA=="
	if protocolLineKind(b64) != "base64" {
		t.Fatalf("got %q", protocolLineKind(b64))
	}
}
