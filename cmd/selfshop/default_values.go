package main

import "time"

var defaultValues = map[string]any{
	"app.runmode":              "prod",
	"app.name":                 "selfshop",
	"app.shutdown.drain_delay": 3 * time.Second,
	"app.shutdown.timeout":     8 * time.Second,

	"log.min_level":            "info",
	"log.format":               "auto",
	"log.sampler.always_level": "warn",
	"log.sampler.interval":     time.Second,
	"log.sampler.first":        100,
	"log.sampler.thereafter":   100,
}
