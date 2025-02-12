package share_test

import (
	"bytes"
	"context"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
	"github.com/ross96D/updater/share/match"
	"github.com/ross96D/updater/share/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// const testPath = "test_path"
const testSysPath = "test_sys_path"

type TestData map[string]io.Reader

func (t TestData) Get(name string) io.ReadCloser {
	return io.NopCloser(t[name])
}

// TODO add tests
func TestUpdateApp(t *testing.T) {
	createRandomData := func(size uint) io.Reader {
		buff := make([]byte, 0, size)
		for i := 0; i < int(size); i++ {
			buff = append(buff, byte(rand.UintN(256)))
		}
		return bytes.NewBuffer(buff)
	}

	cwd, err := os.Getwd()
	require.NoError(t, err)
	testSysPath := filepath.Join(cwd, testSysPath)

	app := configuration.Application{
		Assets: []configuration.Asset{
			{
				Name:       "asset1",
				SystemPath: filepath.Join(testSysPath, "asset1"),
			},
		},
	}

	data := make(map[string]io.Reader)
	data["asset1"] = createRandomData(500)

	err = match.Update(context.Background(), app, match.WithData(TestData(data)))
	require.NoError(t, err)

	for k, v := range app.Assets {
		_, err := os.Stat(v.SystemPath)
		require.NoError(t, err, "%s", k)
	}
}

func TestDependencyCyclicError(t *testing.T) {
	errConfig := `
	port:            11111
	user_secret_key: ""
	user_jwt_expiry: "2m"
	apps: [
		{
			assets_dependency: {
				"asset1": ["asset2"]
				"asset2": ["asset3"]
				"asset3": ["asset1"]
			}
			auth_token: "-"
			service:    "nothing"

			assets: [
				{
					name:        "asset1"
					service:     "-"
					system_path: "-"
				},
				{
					name:        "asset2"
					system_path: "path1"
				},
				{
					name:        "asset3"
					system_path: "path1"
				},
			]
		},
		{
			auth_token: "-"

			assets: [
				{
					name:        "asset2"
					service:     "-"
					system_path: "-"
					cmd: {
						command: "cmd"
						args: ["arg1", "arg2"]
						path: "some/path"
					}
				},
			]
		},
	]
	`
	err := share.ReloadString(errConfig)
	require.Error(t, err)
}

func TestDependencyNoCyclicError(t *testing.T) {
	errConfig := `
	port:            11111
	user_secret_key: ""
	user_jwt_expiry: "2m"
	apps: [
		{
			assets_dependency: {
				"asset1": ["asset2"]
				"asset2": ["asset3", "asset5"]
				"asset7": ["asset3", "asset5", "asset1"]
				"asset5": ["asset3"]
				"asset6": ["asset7", "asset1"]
			}
			auth_token: "-"
			service:    "nothing"

			assets: [
				{
					name:        "asset1"
					system_path: "path1"
				},
				{
					name:        "asset2"
					system_path: "path1"
				},
				{
					name:        "asset3"
					system_path: "path1"
				},
				{
					name:        "asset4"
					system_path: "path1"
				},
				{
					name:        "asset5"
					system_path: "path1"
				},
				{
					name:        "asset6"
					system_path: "path1"
				},
				{
					name:        "asset7"
					system_path: "path1"
				},
			]
		},
		{
			auth_token: "-"

			assets: [
				{
					name:        "asset2"
					service:     "-"
					system_path: "-"
					cmd: {
						command: "cmd"
						args: ["arg1", "arg2"]
						path: "some/path"
					}
				},
			]
		},
	]
	`
	err := share.ReloadString(errConfig)
	require.NoError(t, err)
	// "asset1": ["asset2"]
	// "asset2": ["asset3", "asset5"]
	// "asset7": ["asset3", "asset5", "asset1"]
	// "asset5": ["asset3"]
	// "asset6": ["asset7", "asset1"]
	expected := []configuration.AssetOrder{
		{
			Asset: configuration.Asset{
				Name:       "asset3",
				SystemPath: "path1",
			},
			Independent: true,
		},
		{
			Asset: configuration.Asset{
				Name:       "asset4",
				SystemPath: "path1",
			},
			Independent: true,
		},
		{
			Asset: configuration.Asset{
				Name:       "asset5",
				SystemPath: "path1",
			},
			Independent: false,
		},
		{
			Asset: configuration.Asset{
				Name:       "asset2",
				SystemPath: "path1",
			},
			Independent: false,
		},
		{
			Asset: configuration.Asset{
				Name:       "asset1",
				SystemPath: "path1",
			},
			Independent: false,
		},
		{
			Asset: configuration.Asset{
				Name:       "asset7",
				SystemPath: "path1",
			},
			Independent: false,
		},
		{
			Asset: configuration.Asset{
				Name:       "asset6",
				SystemPath: "path1",
			},
			Independent: false,
		},
	}
	actual := share.Config().Apps[0].AsstesOrder
	require.Equal(t, len(expected), len(actual))
	require.True(t, actual[0].Independent)
	require.True(t, actual[1].Independent)
	for i := 2; i < len(actual); i++ {
		println(actual[i].Name, expected[i].Name)
		require.Equal(t, expected[i], actual[i])
	}
}

func TestReload(t *testing.T) {
	err := share.Init("config_test.cue")
	require.NoError(t, err)
	old := share.Config()

	expected := configuration.Configuration{
		Port:          1234,
		UserSecretKey: "some_key",
		UserJwtExpiry: configuration.Duration(2 * time.Minute),
		Apps:          []configuration.Application{},
		Users:         []configuration.User{},
		BasePath:      share.DefaultPath,
	}
	require.Equal(t, expected, old)

	err = share.Reload("config_test_reload.cue")
	require.NoError(t, err)
	reloaded := share.Config()

	require.NotEqual(t, old, reloaded)

	expected = configuration.Configuration{
		Port:          1234,
		UserSecretKey: "some_key",
		UserJwtExpiry: configuration.Duration(2 * time.Hour),
		Apps: []configuration.Application{
			{
				AuthToken: "auth",
				Assets: []configuration.Asset{
					{
						Name:       "some asset name",
						Service:    "/is/a/path",
						SystemPath: "/is/a/path",
					},
				},
				Command: &configuration.Command{
					Command: "python",
					Args:    []string{"-f", "-s"},
					Timeout: configuration.Duration(5 * time.Minute),
				},
			},
		},
		Users:    []configuration.User{},
		BasePath: share.DefaultPath,
	}
	for i := 0; i < len(expected.Apps); i++ {
		expected.Apps[i].AsstesOrder = reloaded.Apps[i].AsstesOrder
	}
	require.Equal(t, expected, reloaded)
}

func TestConfigPathValidationLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.SkipNow()
	}
	conf := configuration.Configuration{
		BasePath: "/valid/path",
		Apps: []configuration.Application{
			{
				Assets: []configuration.Asset{
					{
						SystemPath: "/app/valid/path",
					},
				},
			},
		},
	}

	err := share.ConfigPathValidation(conf)
	assert.Equal(t, []string{}, err)
}

func TestConfigPathValidationWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.SkipNow()
	}
	conf := configuration.Configuration{
		BasePath: "C:\\valid\\path",
		Apps: []configuration.Application{
			{
				Assets: []configuration.Asset{
					{
						SystemPath: "D:\\app\\valid\\path",
					},
					{
						SystemPath: "D:\\asset\\valid\\path",
					},
				},
			},
		},
	}

	err := share.ConfigPathValidation(conf)
	assert.Equal(t, []string{}, err)
}

func TestPostActionCommand(t *testing.T) {
	app := configuration.Application{
		Command: &configuration.Command{
			Command: "echo",
			Args:    []string{"-n", "test"},
		},
	}

	err := match.NewAppUpdater(
		context.Background(),
		app,
		match.WithData(match.NoData{}),
		match.WithDryRun(true),
	).RunPostAction()

	require.NoError(t, err)
}

func TestUnzip(t *testing.T) {
	t.Run("zip ext", func(t *testing.T) {
		err := utils.Unzip(filepath.Join("unzip_test", "compressed_test.zip"))
		require.NoError(t, err)

		f, err := os.Open(filepath.Join("unzip_test", "compressed_test"))
		require.NoError(t, err)

		t.Cleanup(func() {
			f.Close()
			os.Remove(filepath.Join("unzip_test", "compressed_test"))
		})

		require.NoError(t, err)
		b, err := io.ReadAll(f)
		require.NoError(t, err)

		builder := strings.Builder{}
		for i := 0; i < 50; i++ {
			builder.WriteString("compressed test file with zip\n")
		}

		require.Equal(t, string(b), builder.String())
	})

	t.Run("gz ext", func(t *testing.T) {
		err := utils.Unzip(filepath.Join("unzip_test", "compressed_test.gz"))
		require.NoError(t, err)

		f, err := os.Open(filepath.Join("unzip_test", "compressed_test"))
		require.NoError(t, err)

		t.Cleanup(func() {
			f.Close()
			os.Remove(filepath.Join("unzip_test", "compressed_test"))
		})

		b, err := io.ReadAll(f)
		require.NoError(t, err)

		builder := strings.Builder{}
		for i := 0; i < 50; i++ {
			builder.WriteString(strconv.Itoa(i+1) + " compressed test file with zip\n")
		}

		require.Equal(t, builder.String(), string(b), "lines from actual %d", strings.Count(string(b), "\n"))
	})

	t.Run("tar ext", func(t *testing.T) {
		err := utils.Unzip(filepath.Join("unzip_test", "test.tar"))
		require.NoError(t, err)

		f1, err := os.Open(filepath.Join("unzip_test", "tar.1"))
		require.NoError(t, err)
		f2, err := os.Open(filepath.Join("unzip_test", "tar.2"))
		require.NoError(t, err)

		t.Cleanup(func() {
			f1.Close()
			f2.Close()
			os.Remove(filepath.Join("unzip_test", "tar.1"))
			os.Remove(filepath.Join("unzip_test", "tar.2"))
		})

		b1, err := io.ReadAll(f1)
		require.NoError(t, err)
		b2, err := io.ReadAll(f2)
		require.NoError(t, err)

		assert.Equal(t, "tar.gz.1\n", string(b1))
		assert.Equal(t, "tar.gz.2\n", string(b2))
	})

	t.Run("tar gz ext", func(t *testing.T) {
		err := utils.Unzip(filepath.Join("unzip_test", "test.tar.gz"))
		require.NoError(t, err)

		f1, err := os.Open(filepath.Join("unzip_test", "tar.gz.1"))
		require.NoError(t, err)
		f2, err := os.Open(filepath.Join("unzip_test", "tar.gz.2"))
		require.NoError(t, err)

		t.Cleanup(func() {
			f1.Close()
			f2.Close()
			os.Remove(filepath.Join("unzip_test", "tar.gz.1"))
			os.Remove(filepath.Join("unzip_test", "tar.gz.2"))
		})

		b1, err := io.ReadAll(f1)
		require.NoError(t, err)
		b2, err := io.ReadAll(f2)
		require.NoError(t, err)

		assert.Equal(t, "tar.gz.1\n", string(b1))
		assert.Equal(t, "tar.gz.2\n", string(b2))
	})

	t.Run("gzipDecompress untar Copy() failed unexpected EOF", func(t *testing.T) {
		err := utils.Unzip(filepath.Join("unzip_test", "package.tar.gz"))
		require.NoError(t, err)

		t.Cleanup(func() {
			os.Remove(filepath.Join("unzip_test", "package-lock.json"))
			os.Remove(filepath.Join("unzip_test", "package.json"))
		})
	})
}
