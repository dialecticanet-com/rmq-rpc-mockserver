package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		envFile := "MYAPP_ENV_FROM_FILE=someval1"
		err := os.Setenv("MYAPP_ENV_FROM_ENV", "someval2")
		require.NoError(t, err)

		envFileName := t.TempDir() + "/.env"

		err = os.WriteFile(envFileName, []byte(envFile), 0600)
		require.NoError(t, err)

		type config struct {
			ValueFromFile string `envconfig:"ENV_FROM_FILE" required:"true"`
			ValueFromEnv  string `envconfig:"ENV_FROM_ENV" required:"true"`
			DefaultVal    string `envconfig:"DEFAULT_VAL" default:"default" required:"true"`
		}

		c := &config{}
		err = Load(c, WithEnvFile(envFileName), WithEnvPrefix("MYAPP"))
		require.NoError(t, err)

		assert.Equal(t, "someval1", c.ValueFromFile)
		assert.Equal(t, "someval2", c.ValueFromEnv)
		assert.Equal(t, "default", c.DefaultVal)
	})

	t.Run("file not found", func(t *testing.T) {
		envFileName := t.TempDir() + "/.env"

		type config struct {
			ValueFromFile string `envconfig:"ENV_FROM_FILE" required:"true"`
		}

		c := config{}
		err := Load(&c, WithEnvFile(envFileName))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error is occurred on")
	})

	t.Run("error on loading config", func(t *testing.T) {
		type config struct {
			SomeParameter string `envconfig:"NON_EXISTING_ENV_VAR" required:"true"`
		}

		c := config{}
		err := Load(&c)
		t.Logf("error: %v", err)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "error is occurred on loading config: required key NON_EXISTING_ENV_VAR missing value")

	})
}
