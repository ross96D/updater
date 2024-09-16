package server_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/ross96D/updater/server"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/utils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateEnpoint(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)

	configuration := `
	port:            7432
	user_secret_key: "secret_key"
	user_jwt_expiry: "2h"
	
	apps: [
		{
			auth_token: "identifier-secret-token"
	
			assets: [
				{
					name:        "zipped"
					system_path: "/not/a/path"
					unzip:       true
				},
				{
					name: "service.zipped"
					system_path: "/not/a/path"
					service: "service-name"
					unzip:       true
				},
				{
					name: "service.zipped.command"
					system_path: "/not/a/path"
					service: "service-with-command-name"
					unzip:       true
					cmd: {
						command: "npm"
						args: ["run", "dev"]
						path: "not/a/path"
					}
				},
				{
					name: "dumb_test"
					system_path: "` + filepath.Join(dir, "server_data_test", "dumb_file.txt") + `"
					cmd: {
						command: "echo"
						args: ["test", "command", "output"]
					}
				}
			]
		},
	]
	`
	log.Logger = log.Logger.Output(io.Discard)

	type testAsset struct {
		name string
		data string
	}
	type expectError int
	const (
		noerror = iota
		nofatal
		fatal
	)
	type testData struct {
		name        string
		assets      []testAsset
		expectError expectError
		useImpl     bool
	}

	err = share.ReloadString(configuration)
	require.NoError(t, err)

	data := []testData{
		{
			name: "update with out match",
			assets: []testAsset{
				{
					name: "inexistent",
					data: "no data",
				},
			},
			expectError: nofatal,
		},
		{
			name: "update match all",
			assets: []testAsset{
				{
					name: "zipped",
					data: "-",
				},
				{
					name: "service.zipped",
					data: "-",
				},
				{
					name: "service.zipped.command",
					data: "-",
				},
				{
					name: "dumb_test",
					data: "-",
				},
			},
			expectError: noerror,
		},
		{
			name: "update no fatal with impl",
			assets: []testAsset{
				{
					name: "dumb_test",
					data: "dumb-file2\n",
				},
			},
			useImpl:     true,
			expectError: nofatal,
		},
	}
	t.Cleanup(func() {
		err = os.Rename(
			filepath.Join(dir, "server_data_test", "dumb_file.txt.old"),
			filepath.Join(dir, "server_data_test", "dumb_file.txt"),
		)
		if err != nil {
			panic(err)
		}
	})

	r := regexp.MustCompile(`\d{1,2}:\d{1,2}[A,P]M (?<Level>[^\s]+)`)

	for _, data := range data {
		t.Run(data.name, func(t *testing.T) {
			buff := &bytes.Buffer{}
			// append assets to multipart reader
			multipartWriter := multipart.NewWriter(buff)
			for _, asset := range data.assets {
				fieldW, err := multipartWriter.CreateFormFile(asset.name, asset.name)
				require.NoError(t, err)
				var buff *bytes.Buffer = &bytes.Buffer{}
				buff.Write([]byte(asset.data))
				_, err = io.Copy(fieldW, buff)
				require.NoError(t, err)
			}
			multipartWriter.Close()

			req := httptest.NewRequest(http.MethodPost, "/update", buff)
			req.Header.Set("Authorization", "identifier-secret-token")
			if !data.useImpl {
				req.Header.Set("dry-run", "true")
			}
			req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

			w := httptest.NewRecorder()
			server.New("", "").TestServeHTTP(w, req)

			res := w.Result()
			body, err := io.ReadAll(res.Body)
			assert.Equal(t, 200, res.StatusCode, string(body))
			require.NoError(t, err)
			defer res.Body.Close()

			lines := bytes.Split(body, []byte("\n"))
			last := lines[len(lines)-2]
			require.True(t, r.Match(utils.StripAnsiBytes(last)))
			level := r.FindSubmatch(utils.StripAnsiBytes(last))[r.SubexpIndex("Level")]
			switch data.expectError {
			case noerror:
				assert.NotEqual(t, "ERR", string(level))
				assert.NotEqual(t, "WRN", string(level))

			case nofatal:
				assert.NotEqual(t, "ERR", string(level))

			case fatal:
				assert.Equal(t, "ERR", string(level))

			}
		})
	}
}

func GetData(sizeInBytes uint64) io.Reader {
	buff := bytes.NewBuffer([]byte{})
	for i := 0; i < int(sizeInBytes); i++ {
		buff.WriteByte(byte(rand.N(256)))
	}
	return buff
}

func TestParseMultipartForm(t *testing.T) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	// create a new form-data header name data and filename data.txt
	dataPart, err := writer.CreateFormFile("data", "data")
	require.NoError(t, err)

	hasher := crc32.New(crc32.IEEETable)

	// copy file content into multipart section dataPart
	f := GetData(30 << 20)
	f = io.TeeReader(f, hasher)

	_, err = io.Copy(dataPart, f)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	dataHash := hex.EncodeToString(hasher.Sum(nil))

	// create HTTP request & response
	r, err := http.NewRequest(http.MethodPost, "/files", body)
	require.NoError(t, err)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	func(w http.ResponseWriter, r *http.Request) {
		data, err := server.ParseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		reader := data.Get("data")
		if reader == nil {
			http.Error(w, "no data found", 500)
			return
		}

		hasher := crc32.New(crc32.IEEETable)
		io.Copy(hasher, data.Get("data"))                                            //nolint: errcheck
		w.Write([]byte(fmt.Sprintf("data:%s", hex.EncodeToString(hasher.Sum(nil))))) //nolint: errcheck
	}(w, r)

	// validate
	message, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Result().StatusCode, "%s", message)

	require.Equal(t, string(message), fmt.Sprintf("data:%s", dataHash))
}
