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
	"regexp"
	"testing"

	"github.com/ross96D/updater/server"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/utils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerEnpoints(t *testing.T) {
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
					service: "service name"
					unzip:       true
				},
				{
					name: "service.zipped.command"
					system_path: "/not/a/path"
					service: "service with command name"
					unzip:       true
					cmd: {
						command: "npm"
						args: ["run", "dev"]
						path: "not/a/path"
					}
				},
			]
		},
	]
	
	`
	log.Logger = log.Logger.Output(io.Discard)

	t.Run("/update with out match", func(t *testing.T) {
		err := share.ReloadString(configuration)
		require.NoError(t, err)

		buff := &bytes.Buffer{}
		multipartWriter := multipart.NewWriter(buff)
		{ // Create Form field
			fieldW, err := multipartWriter.CreateFormFile("ctm_web", "ctm_web.tar.gz")
			require.NoError(t, err)
			var buff *bytes.Buffer = &bytes.Buffer{}
			buff.Write([]byte("testing"))
			_, err = io.Copy(fieldW, buff)
			require.NoError(t, err)
		}
		multipartWriter.Close()
		req := httptest.NewRequest(http.MethodPost, "/update", buff)
		req.Header.Add("Authorization", "identifier-secret-token")
		req.Header.Add("dry-run", "true")
		req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

		w := httptest.NewRecorder()

		server.New("", "").TestServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, 200, res.StatusCode)

		println("response")
		println(string(data))
		lines := bytes.Split(data, []byte("\n"))
		last := lines[len(lines)-2]
		r := regexp.MustCompile(`\d{1,2}:\d{1,2}[A,P]M (?<Level>[^\s]+)`)
		require.True(t, r.Match(utils.StripAnsiBytes(last)))
		level := r.FindSubmatch(utils.StripAnsiBytes(last))[r.SubexpIndex("Level")]
		assert.NotEqual(t, "ERR", string(level))
	})
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
		_, _ = io.Copy(hasher, data.Get("data"))
		_, _ = w.Write([]byte(fmt.Sprintf("data:%s", hex.EncodeToString(hasher.Sum(nil)))))
	}(w, r)

	// validate
	message, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Result().StatusCode, "%s", message)

	require.Equal(t, string(message), fmt.Sprintf("data:%s", dataHash))
}
