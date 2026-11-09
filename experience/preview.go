package experience

import (
	"fmt"
)

// LiveSession describes a connected session for reload impact analysis.
type LiveSession struct {
	ID     string
	Scene  SceneID
	Spaces []SpaceID
}

// StrandedSession is a session that would be orphaned by a proposed wire.
type StrandedSession struct {
	ID     string `json:"id"`
	Scene  string `json:"scene"`
	Reason string `json:"reason"`
}

// ReloadPreview summarizes whether a wire could replace the live experience graph.
type ReloadPreview struct {
	OK               bool              `json:"ok"`
	Errors           []string          `json:"errors,omitempty"`
	StrandedSessions []StrandedSession `json:"stranded_sessions,omitempty"`
	Warnings         []string          `json:"warnings,omitempty"`
}

// PreviewReload validates proposed wiring against live sessions (same scene and space rule as runtime hot-reload).
func PreviewReload(current *Experience, proposed Wire, live []LiveSession) ReloadPreview {
	var out ReloadPreview
	if err := ValidateWire(proposed); err != nil {
		out.Errors = append(out.Errors, err.Error())
		return out
	}
	if proposed.Root == "" {
		out.Errors = append(out.Errors, "portico: wire root is required")
		return out
	}
	if _, ok := proposed.Scenes[proposed.Root]; !ok {
		out.Errors = append(out.Errors, fmt.Sprintf("portico: wire root scene %q not in scenes", proposed.Root))
		return out
	}
	proposedSpaces := make(map[string]struct{}, len(proposed.Spaces))
	for _, name := range proposed.Spaces {
		proposedSpaces[name] = struct{}{}
	}
	proposedScenes := make(map[string]struct{}, len(proposed.Scenes))
	for name := range proposed.Scenes {
		proposedScenes[name] = struct{}{}
	}
	if current != nil && len(proposed.Assets) > 0 {
		for _, name := range proposed.Assets {
			if _, ok := current.Asset(name); !ok {
				out.Warnings = append(out.Warnings, fmt.Sprintf("asset %q not loaded in current experience", name))
			}
		}
	}
	for _, sess := range live {
		scene := string(sess.Scene)
		if _, ok := proposedScenes[scene]; !ok {
			out.StrandedSessions = append(out.StrandedSessions, StrandedSession{
				ID:     sess.ID,
				Scene:  scene,
				Reason: "scene removed from proposed experience",
			})
		}
		for _, sp := range sess.Spaces {
			if _, ok := proposedSpaces[string(sp)]; !ok {
				out.StrandedSessions = append(out.StrandedSessions, StrandedSession{
					ID:     sess.ID,
					Scene:  scene,
					Reason: fmt.Sprintf("session in space %q removed from proposed experience", sp),
				})
			}
		}
	}
	out.OK = len(out.Errors) == 0 && len(out.StrandedSessions) == 0
	return out
}
