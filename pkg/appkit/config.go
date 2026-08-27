package appkit

type Config struct {
	Runmode Runmode `koanf:"runmode"`
	Name    string  `koanf:"name"`
}

func (c Config) IsDevmod() bool { return c.Runmode == RunmodeDev }
