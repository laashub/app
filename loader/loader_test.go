package loader

import (
	"io/ioutil"
	"testing"

	"github.com/docker/app/internal"
	"github.com/docker/app/types"
	"github.com/docker/docker/pkg/archive"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
	"gotest.tools/fs"
)

const (
	metadata = `name: my-app
version: 1.0.0
`
	compose = `version: "3.1"
services:
  web:
    image: nginx
`
	params = `foo: bar
`
)

func TestLoadFromDirectory(t *testing.T) {
	dir := fs.NewDir(t, "my-app",
		fs.WithFile(internal.MetadataFileName, metadata),
		fs.WithFile(internal.ParametersFileName, params),
		fs.WithFile(internal.ComposeFileName, compose),
	)
	defer dir.Remove()
	app, err := LoadFromDirectory(dir.Path())
	assert.NilError(t, err)
	assert.Assert(t, app != nil)
	assert.Assert(t, is.Equal(app.Path, dir.Path()))
	assertAppContent(t, app)
}

func TestLoadFromDirectoryDeprecatedSettings(t *testing.T) {
	dir := fs.NewDir(t, "my-app",
		fs.WithFile(internal.MetadataFileName, metadata),
		fs.WithFile(internal.DeprecatedSettingsFileName, params),
		fs.WithFile(internal.ComposeFileName, compose),
	)
	defer dir.Remove()
	app, err := LoadFromDirectory(dir.Path())
	assert.Assert(t, app == nil)
	assert.ErrorContains(t, err, "\"settings.yml\" has been deprecated in favor of \"parameters.yml\"; please rename \"settings.yml\" to \"parameters.yml\"")
}

func TestLoadFromTarInexistent(t *testing.T) {
	_, err := LoadFromTar("any-tar.tar")
	assert.ErrorContains(t, err, "open any-tar.tar")
}

func TestLoadFromTar(t *testing.T) {
	myapp := createAppTar(t)
	defer myapp.Remove()
	app, err := LoadFromTar(myapp.Path())
	assert.NilError(t, err)
	assert.Assert(t, app != nil)
	assert.Assert(t, is.Equal(app.Path, myapp.Path()))
	assertAppContent(t, app)
}

func createAppTar(t *testing.T) *fs.File {
	t.Helper()
	dir := fs.NewDir(t, "my-app",
		fs.WithFile(internal.MetadataFileName, metadata),
		fs.WithFile(internal.ParametersFileName, params),
		fs.WithFile(internal.ComposeFileName, compose),
	)
	defer dir.Remove()
	r, err := archive.TarWithOptions(dir.Path(), &archive.TarOptions{
		Compression: archive.Uncompressed,
	})
	assert.NilError(t, err)
	data, err := ioutil.ReadAll(r)
	assert.NilError(t, err)
	return fs.NewFile(t, "app", fs.WithBytes(data))
}

func assertContentIs(t *testing.T, actual []byte, expected string) {
	t.Helper()
	assert.Assert(t, is.Equal(string(actual), expected))
}

func assertAppContent(t *testing.T, app *types.App) {
	assert.Assert(t, is.Len(app.ParametersRaw(), 1))
	assertContentIs(t, app.ParametersRaw()[0], params)
	assert.Assert(t, is.Len(app.Composes(), 1))
	assertContentIs(t, app.Composes()[0], compose)
	assertContentIs(t, app.MetadataRaw(), metadata)
}
