// Reference consumer inside this module: RecipeCommunity + ServeDev.
// SSH sites outside this repo import github.com/MrBlindMouse/Portico; they are not added under examples/.
package main

import (
	"log"
	"os"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/scaffold"
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

func build() (*experience.Experience, error) {
	return portico.New().
		Merge(scaffold.RecipeCommunity()).
		Build()
}
