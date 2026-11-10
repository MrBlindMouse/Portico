// Capability probe inside this module. Sites do not copy it.
// SSH sites outside this repo import github.com/MrBlindMouse/Portico; they are not added under examples/.
package main

import (
	"log"
	"os"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/internal/plans"
)

func main() {
	exp, err := build()
	if err != nil {
		log.Fatal(err)
	}
	dataDir := ".portico-data"
	if v := os.Getenv("PORTICO_DATA"); v != "" {
		dataDir = v
	}
	opts := portico.Options{DataDir: dataDir}
	if v := os.Getenv("PORTICO_LOG_LEVEL"); v != "" {
		opts.LogLevel = v
	}
	if err := portico.ServeDev(exp, opts); err != nil {
		log.Fatal(err)
	}
}

func build() (*installation.Installation, error) {
	return portico.New().
		Merge(plans.Guestbook()).
		Merge(plans.Chat()).
		Merge(plans.Landing(map[string]components.SceneID{
			"guestbook": "guestbook",
			"chat":      "chat",
		})).
		Build()
}
