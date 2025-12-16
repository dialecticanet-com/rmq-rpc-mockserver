package config

// Option is a configuration option.
type Option func(*configOpts)

type configOpts struct {
	envFile   string
	envPrefix string
}

func defaultOptions() *configOpts {
	return &configOpts{
		envFile: "",
	}
}

// WithEnvFile sets the path to the environment file that will be loaded to ENV variables automatically.
func WithEnvFile(envFile string) Option {
	return func(c *configOpts) {
		c.envFile = envFile
	}
}

// WithEnvPrefix sets the prefix for the environment variables.
// e.g. if all environment variables are prefixed with "MYAPP_",
// then you can set the prefix to "MYAPP" and avoid adding it all `envconfig` struct tags.
func WithEnvPrefix(envPrefix string) Option {
	return func(c *configOpts) {
		c.envPrefix = envPrefix
	}
}
