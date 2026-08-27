package config

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/joho/godotenv"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

type Validator interface{ Validate() error }

const delim = "."

func New[T any](
	envPrefix string, defaultValues map[string]any,
) (*T, error) {
	var zero T
	rt := reflect.TypeOf(zero)
	if rt == nil {
		return nil, fmt.Errorf("type parameter T must be a struct")
	}
	for rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type parameter T must be a struct, got %s", rt.Kind())
	}
	k := koanf.New(delim)

	envPrefix = strings.TrimRight(envPrefix, "_") + "_"
	envPrefix = strings.ToUpper(envPrefix)

	_ = k.Load(confmap.Provider(defaultValues, delim), nil) //nolint:errcheck // The confmap provider reads
	// from an in-memory map and cannot produce an error.

	_ = godotenv.Load() //nolint:errcheck // A missing .env file is
	// expected; loading it is best-effort only.

	//nolint:errcheck // env.Provider only exposes process environment variables and cannot fail.
	_ = k.Load(env.Provider(delim, env.Opt{
		Prefix:        envPrefix,
		TransformFunc: envKeyTransform(envPrefix, delim),
	}), nil)

	var conf T
	if err := k.UnmarshalWithConf("", &conf, koanf.UnmarshalConf{
		DecoderConfig: decoderConfig(),
	}); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	if v, ok := any(conf).(Validator); ok {
		if err := v.Validate(); err != nil {
			return nil, fmt.Errorf("semantic validation: %w", err)
		}
	}
	return &conf, nil
}

func envKeyTransform(prefix, delim string) func(string, string) (string, any) {
	return func(k, v string) (string, any) {
		k = strings.TrimPrefix(k, prefix)
		k = strings.ToLower(k)
		k = strings.TrimLeft(k, "_")
		k = strings.ReplaceAll(k, "__", delim)
		return k, v
	}
}

func decoderConfig() *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		IgnoreUntaggedFields: false,
		ErrorUnused:          true,
		WeaklyTypedInput:     false,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToTimeHookFunc(time.RFC3339),
			mapstructure.StringToBasicTypeHookFunc(),
		),
	}
}
