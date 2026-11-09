// Reference Portico site: merge scaffolds, then ServeDev. Copy this package as your own main.
package main

import (
	"log"
	"os"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/scaffold"
)

func main() {
	exp, err := portico.New().
		Merge(scaffold.Guestbook()).
		Merge(scaffold.Chat()).
		Merge(scaffold.Landing(map[string]experience.SceneID{
			"guestbook": "guestbook",
			"chat":      "chat",
		})).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	dataDir := ".portico-data"
	if v := os.Getenv("PORTICO_DATA"); v != "" {
		dataDir = v
	}
	if err := portico.ServeDev(exp, portico.Options{DataDir: dataDir}); err != nil {
		log.Fatal(err)
	}
}
