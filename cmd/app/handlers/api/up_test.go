package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/lib/oapi"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/onmetal-dev/metal/lib/store/mock"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.jetify.com/typeid"
)

func TestUp(t *testing.T) {
	envId := typeid.Must(typeid.WithPrefix("env"))
	appId := typeid.Must(typeid.WithPrefix("app"))
	teamId := typeid.Must(typeid.WithPrefix("team"))

	t.Run("400 responses", func(t *testing.T) {
		testCases := []struct {
			name    string
			envId   string
			appId   string
			archive bool
			errMsg  string
		}{
			{"missing archive", envId.String(), appId.String(), false, "archive is required"},
			{"missing env_id", "", appId.String(), true, "env_id is required"},
			{"missing app_id", envId.String(), "", true, "app_id is required"},
			{"invalid env_id", "invalid_env_id", appId.String(), true, "invalid env_id"},
			{"invalid app_id", envId.String(), "invalid_app_id", true, "invalid app_id"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				api := newTestAPI()
				ctx := middleware.WithApiToken(context.Background(), store.ApiToken{TeamId: teamId.String()})
				req := oapi.UpRequestObject{
					Body: createMultipartBody(t, tc.envId, tc.appId, tc.archive),
				}

				resp, err := api.Up(ctx, req)
				require.NoError(t, err)

				badReq, ok := resp.(oapi.Up400JSONResponse)
				require.True(t, ok, "Expected 400 response")
				assert.Contains(t, badReq.Error, tc.errMsg)
			})
		}
	})

	t.Run("app not found", func(t *testing.T) {
		api := newTestAPI()
		api.appStore.(*mock.AppStoreMock).On("Get", testifymock.Anything, appId.String()).Return(store.App{}, store.ErrAppNotFound)

		ctx := middleware.WithApiToken(context.Background(), store.ApiToken{TeamId: teamId.String()})
		req := oapi.UpRequestObject{
			Body: createMultipartBody(t, envId.String(), appId.String(), true),
		}

		resp, err := api.Up(ctx, req)
		require.NoError(t, err)

		internalErr, ok := resp.(oapi.Up500JSONResponse)
		require.True(t, ok, "Expected 500 response")
		assert.Contains(t, internalErr.Error, "failed to get app")
	})

	t.Run("env not found", func(t *testing.T) {
		api := newTestAPI()
		api.appStore.(*mock.AppStoreMock).On("Get", testifymock.Anything, appId.String()).Return(store.App{TeamId: teamId.String()}, nil)
		api.deploymentStore.(*mock.DeploymentStoreMock).On("GetEnv", envId.String()).Return(store.Env{}, store.ErrEnvNotFound)

		ctx := middleware.WithApiToken(context.Background(), store.ApiToken{TeamId: teamId.String()})
		req := oapi.UpRequestObject{
			Body: createMultipartBody(t, envId.String(), appId.String(), true),
		}

		resp, err := api.Up(ctx, req)
		require.NoError(t, err)

		internalErr, ok := resp.(oapi.Up500JSONResponse)
		require.True(t, ok, "Expected 500 response")
		assert.Contains(t, internalErr.Error, "failed to get env")
	})

	t.Run("success case", func(t *testing.T) {
		api := newTestAPI()
		api.appStore.(*mock.AppStoreMock).On("Get", testifymock.Anything, appId.String()).Return(store.App{TeamId: teamId.String()}, nil)
		api.deploymentStore.(*mock.DeploymentStoreMock).On("GetEnv", envId.String()).Return(store.Env{TeamId: teamId.String()}, nil)

		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp("", "test-up-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Write some test files
		testFiles := map[string]string{
			"file1.txt":        "Content of file 1",
			"file2.txt":        "Content of file 2",
			"subdir/file3.txt": "Content of file 3 in subdirectory",
		}
		for path, content := range testFiles {
			fullPath := filepath.Join(tempDir, path)
			err := os.MkdirAll(filepath.Dir(fullPath), 0755)
			require.NoError(t, err)
			err = os.WriteFile(fullPath, []byte(content), 0644)
			require.NoError(t, err)
		}

		ctx := middleware.WithApiToken(context.Background(), store.ApiToken{TeamId: teamId.String()})
		req := oapi.UpRequestObject{
			Body: createMultipartBodyWithFiles(t, envId.String(), appId.String(), tempDir),
		}

		resp, err := api.Up(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(oapi.Up200JSONResponse)
		require.True(t, ok, "Expected 200 response")
		assert.NotNil(t, successResp.Message)
		assert.Contains(t, *successResp.Message, "Archive successfully uploaded and extracted")
	})
}

func createMultipartBody(t *testing.T, envId, appId string, includeArchive bool) *multipart.Reader {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if envId != "" {
		err := writer.WriteField("env_id", envId)
		require.NoError(t, err)
	}

	if appId != "" {
		err := writer.WriteField("app_id", appId)
		require.NoError(t, err)
	}

	if includeArchive {
		part, err := writer.CreateFormFile("archive", "archive.tar.gz")
		require.NoError(t, err)
		_, err = part.Write([]byte("mock archive content"))
		require.NoError(t, err)
	}

	err := writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "", &body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	reader, err := req.MultipartReader()
	require.NoError(t, err)

	return reader
}

func createMultipartBodyWithFiles(t *testing.T, envId, appId, dirPath string) *multipart.Reader {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	err := writer.WriteField("env_id", envId)
	require.NoError(t, err)

	err = writer.WriteField("app_id", appId)
	require.NoError(t, err)

	// Create a tar.gz archive of the directory
	archiveBuffer := &bytes.Buffer{}
	err = createTarGz(dirPath, archiveBuffer)
	require.NoError(t, err)

	part, err := writer.CreateFormFile("archive", "archive.tar.gz")
	require.NoError(t, err)
	_, err = io.Copy(part, archiveBuffer)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "", &body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	reader, err := req.MultipartReader()
	require.NoError(t, err)

	return reader
}

func createTarGz(sourceDir string, output io.Writer) error {
	gzipWriter := gzip.NewWriter(output)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
