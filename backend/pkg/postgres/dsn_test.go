package postgres

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDSNFromEnvEscapesCredentials(t *testing.T) {
	t.Setenv("POSTGRES_USER", "app")
	t.Setenv("POSTGRES_PASSWORD", "p@ss:w/rd#?")
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_DB", "pins")

	cfg, err := pgxpool.ParseConfig(DSNFromEnv())
	require.NoError(t, err)

	assert.Equal(t, "app", cfg.ConnConfig.User)
	assert.Equal(t, "p@ss:w/rd#?", cfg.ConnConfig.Password)
	assert.Equal(t, "localhost", cfg.ConnConfig.Host)
	assert.Equal(t, uint16(5432), cfg.ConnConfig.Port)
	assert.Equal(t, "pins", cfg.ConnConfig.Database)
}
