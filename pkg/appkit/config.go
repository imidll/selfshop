package appkit

import "time"

type Config struct {
	Runmode  Runmode        `koanf:"runmode"`
	Name     string         `koanf:"name"`
	Shutdown ShutdownConfig `koanf:"shutdown"`
}

type ShutdownConfig struct {
	Timeout    time.Duration `koanf:"timeout"`
	DrainDelay DrainDelay    `koanf:"drain_delay"`
}

func (c Config) IsDevmod() bool { return c.Runmode == RunmodeDev }

func (c ShutdownConfig) TotalTimeout() time.Duration {
	return c.Timeout + c.DrainDelay.Duration()
}
