package inspect

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
)

// Snapshot is the top-level inspect payload.
type Snapshot struct {
	Experience ExperienceSnap `json:"experience"`
	Sessions   []SessionSnap  `json:"sessions"`
	Spaces     map[string]SpaceSnap `json:"spaces"`
}

// ExperienceSnap summarizes the loaded definition.
type ExperienceSnap struct {
	Root         string   `json:"root"`
	Hash         string   `json:"hash"` // scene id set fingerprint (not full definition hash)
	Scenes       []string `json:"scenes"`
}

// SpaceSnap is a read-only space summary for inspection.
type SpaceSnap struct {
	MemberCount int      `json:"member_count"`
	Keys        []string `json:"keys"`
}

// SessionSnap describes one live session.
type SessionSnap struct {
	ID            string            `json:"id"`
	Scene         string            `json:"scene"`
	Identity      string            `json:"identity"`
	Fingerprint   string            `json:"fingerprint"`
	Inputs        map[string]string `json:"inputs"`
	RenderHash    string            `json:"render_hash"`
	RenderSize    int               `json:"render_size"`
	RenderPreview string            `json:"render_preview"`
	Spaces        []string          `json:"spaces"`
}

// Handler serves JSON inspection endpoints.
type Handler struct {
	Engine *runtime.Engine
}

// NewHandler creates an inspect HTTP handler mux.
func NewHandler(eng *runtime.Engine) http.Handler {
	h := &Handler{Engine: eng}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /inspect/snapshot", h.snapshot)
	mux.HandleFunc("GET /inspect/sessions", h.sessions)
	mux.HandleFunc("GET /inspect/session/{id}", h.session)
	mux.HandleFunc("POST /inspect/validate", h.validate)
	mux.HandleFunc("GET /inspect/identity/{fingerprint}", h.identity)
	mux.HandleFunc("GET /inspect/spaces", h.spaces)
	return mux
}

func (h *Handler) snapshot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.buildSnapshot())
}

func (h *Handler) sessions(w http.ResponseWriter, r *http.Request) {
	snap := h.buildSnapshot()
	writeJSON(w, snap.Sessions)
}

func (h *Handler) session(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, s := range h.buildSnapshot().Sessions {
		if s.ID == id {
			writeJSON(w, s)
			return
		}
	}
	http.NotFound(w, r)
}

func (h *Handler) validate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var wire experience.Wire
	if len(body) == 0 {
		writeJSON(w, map[string]any{"ok": true, "note": "empty body; send experience.Wire JSON to dry-run wiring"})
		return
	}
	if err := json.Unmarshal(body, &wire); err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if err := experience.ValidateWire(wire); err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (h *Handler) identity(w http.ResponseWriter, r *http.Request) {
	fp := r.PathValue("fingerprint")
	store := h.Engine.Applier().Store()
	if store == nil {
		writeJSON(w, map[string]any{"error": "no store"})
		return
	}
	id := identity.ForFingerprint(fp)
	keys, err := store.ListKeys(id)
	if err != nil {
		writeJSON(w, map[string]any{"error": err.Error()})
		return
	}
	vals := make(map[string]any, len(keys))
	for _, k := range keys {
		v, ok, err := store.Get(id, k)
		if err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		if ok {
			vals[k] = v
		}
	}
	writeJSON(w, map[string]any{"fingerprint": fp, "values": vals})
}

func (h *Handler) spaces(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, spaceSnaps(h.Engine.SpaceSnapshots()))
}

func spaceSnaps(in map[string]runtime.SpaceSnapshot) map[string]SpaceSnap {
	out := make(map[string]SpaceSnap, len(in))
	for id, s := range in {
		out[id] = SpaceSnap{MemberCount: s.MemberCount, Keys: s.Keys}
	}
	return out
}

func (h *Handler) buildSnapshot() Snapshot {
	exp := h.Engine.Experience()
	scenes := make([]string, 0, len(exp.Scenes()))
	for _, id := range exp.Scenes() {
		scenes = append(scenes, string(id))
	}
	hash := sceneSetHash(exp)
	out := Snapshot{
		Experience: ExperienceSnap{
			Root:   string(exp.Root()),
			Hash:   hash,
			Scenes: scenes,
		},
		Spaces: spaceSnaps(h.Engine.SpaceSnapshots()),
	}
	for _, sess := range h.Engine.Sessions() {
		sc, _ := exp.Scene(sess.CurrentScene())
		inputs := make(map[string]string)
		for k, v := range sc.Inputs {
			inputs[k] = string(v)
		}
		_, _ = h.Engine.RenderText(sess)
		text, size := sess.LastRender()
		preview := text
		const maxPreview = 240
		if len(preview) > maxPreview {
			preview = preview[:maxPreview] + "…"
		}
		spaceIDs := make([]string, 0, len(sess.Spaces()))
		for _, sp := range sess.Spaces() {
			spaceIDs = append(spaceIDs, string(sp))
		}
		out.Sessions = append(out.Sessions, SessionSnap{
			ID:            sess.ID(),
			Scene:         string(sess.CurrentScene()),
			Identity:      sess.Identity().String(),
			Fingerprint:   sess.Identity().Fingerprint(),
			Inputs:        inputs,
			RenderHash:    renderHash(text),
			RenderSize:    size,
			RenderPreview: preview,
			Spaces:        spaceIDs,
		})
	}
	return out
}

func renderHash(text string) string {
	if text == "" {
		return ""
	}
	h := sha256.Sum256([]byte(text))
	return hex.EncodeToString(h[:])[:16]
}

func sceneSetHash(exp *experience.Experience) string {
	h := sha256.New()
	for _, id := range exp.Scenes() {
		h.Write([]byte(string(id)))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(w, `{"error":%q}`, err.Error())
	}
}
