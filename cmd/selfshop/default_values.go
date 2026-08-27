package main

import "time"

var defaultValues = map[string]any{
	"app.runmode": "prod",
	"app.name":    "selfshop",

	"log.min_level":            "info",
	"log.format":               "auto",
	"log.sampler.always_level": "warn",
	"log.sampler.interval":     time.Second,
	"log.sampler.first":        100,
	"log.sampler.thereafter":   100,
}
