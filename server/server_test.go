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
	"testing"

	"github.com/ross96D/updater/server"
	"github.com/stretchr/testify/require"
)

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
