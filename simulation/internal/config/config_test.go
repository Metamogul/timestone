package config

import (
	"github.com/metamogul/timestone/simulation/config"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConvert(t *testing.T) {
	t.Parallel()

	before := config.Before{
		Interval: -1,
		Tags:     []string{"test"},
	}

	beforeInternal := Convert(before, time.Time{})
	require.Equal(t, before.Tags, beforeInternal.Tags)
	require.Equal(t, time.Time{}.Add(before.Interval), beforeInternal.Time)
}
