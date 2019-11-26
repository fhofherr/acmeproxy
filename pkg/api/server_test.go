package api_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/api"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestStartServer(t *testing.T) {
	testsupport.SkipIfPebbleDisabled(t)

	fx := api.NewTestFixture(t)
	defer fx.Close()
	fx.MustStartServer(t)
	defer fx.Server.Shutdown(context.Background()) // nolint

	assert.FileExists(t, filepath.Join(fx.DataDir, "acmeproxy.db"), "database file not created")
}

func TestFailsIfDBCannotBeOpened(t *testing.T) {
	testsupport.SkipIfPebbleDisabled(t)

	fx := api.NewTestFixture(t)
	defer fx.Close()

	// DataDir exists but is not writable
	err := os.Mkdir(fx.DataDir, 0444)
	if !assert.NoError(t, err) {
		return
	}
	err = fx.Server.Start()
	assert.Error(t, err)
}

func TestCannotStartServerTwice(t *testing.T) {
	testsupport.SkipIfPebbleDisabled(t)

	fx := api.NewTestFixture(t)
	defer fx.Close()

	fx.MustStartServer(t)
	defer fx.Server.Shutdown(context.Background()) // nolint

	err := fx.Server.Start()
	assert.True(t, errors.Is(err, errors.New("already started")))
}

func TestCannotReStartAStoppedServer(t *testing.T) {
	testsupport.SkipIfPebbleDisabled(t)

	fx := api.NewTestFixture(t)
	defer fx.Close()

	fx.MustStartServer(t)
	err := fx.Server.Shutdown(context.Background())
	assert.NoError(t, err)

	err = fx.Server.Start()
	assert.True(t, errors.Is(err, errors.New("already started")))
}

func TestShutdownOfAnUnstartedServerHasNoEffect(t *testing.T) {
	testsupport.SkipIfPebbleDisabled(t)

	fx := api.NewTestFixture(t)
	defer fx.Close()

	err := fx.Server.Shutdown(context.Background())
	assert.True(t, errors.Is(err, errors.New("not started")))
}
