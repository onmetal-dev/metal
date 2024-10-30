package api

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/oapi"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/samber/lo"
	"go.jetify.com/typeid"
)

func joinErrors(errors []error) string {
	return strings.Join(lo.Map(errors, func(err error, _ int) string {
		return err.Error()
	}), ", ")
}

func (a api) Up(ctx context.Context, request oapi.UpRequestObject) (oapi.UpResponseObject, error) {
	token := middleware.MustGetApiToken(ctx)

	// Create a temporary file for the archive
	tempFile, err := os.CreateTemp("", "archive-*.tar.gz")
	if err != nil {
		return oapi.Up500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: "failed to create temporary file"}}, nil
	}
	fmt.Println("DEBUG: uploaded archive path", tempFile.Name())
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	var envIdBytes, appIdBytes []byte
	var archiveReceived bool
	for {
		part, err := request.Body.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return oapi.Up400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: fmt.Sprintf("failed to read form part: %s", err)}}, nil
		}

		switch part.FormName() {
		case "env_id":
			envIdBytes, err = io.ReadAll(part)
		case "app_id":
			appIdBytes, err = io.ReadAll(part)
		case "archive":
			_, err = io.Copy(tempFile, part)
			archiveReceived = true
		default:
			return oapi.Up400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: fmt.Sprintf("invalid form part: %s", part.FormName())}}, nil
		}
		if err != nil {
			return oapi.Up500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
		}
	}

	var validationErrors []error
	if !archiveReceived {
		validationErrors = append(validationErrors, errors.New("archive is required"))
	}
	if len(envIdBytes) == 0 {
		validationErrors = append(validationErrors, errors.New("env_id is required"))
	}
	if len(appIdBytes) == 0 {
		validationErrors = append(validationErrors, errors.New("app_id is required"))
	}
	if len(validationErrors) > 0 {
		return oapi.Up400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: joinErrors(validationErrors)}}, nil
	}

	// Validate env_id and app_id
	envId, err := typeid.FromString(string(envIdBytes))
	if err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("invalid env_id: %s", err))
	} else if envId.Prefix() != "env" {
		validationErrors = append(validationErrors, fmt.Errorf("invalid env_id: %s", envId.String()))
	}
	appId, err := typeid.FromString(string(appIdBytes))
	if err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("invalid app_id: %s", err))
	} else if appId.Prefix() != "app" {
		validationErrors = append(validationErrors, fmt.Errorf("invalid app_id: %s", appId.String()))
	}
	if len(validationErrors) > 0 {
		return oapi.Up400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: joinErrors(validationErrors)}}, nil
	}

	app, err := a.appStore.Get(ctx, appId.String())
	if err != nil {
		return oapi.Up500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: fmt.Sprintf("failed to get app: %s", err)}}, nil
	}
	if app.TeamId != token.TeamId {
		return oapi.Up400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: "app does not belong to team"}}, nil
	}

	env, err := a.deploymentStore.GetEnv(envId.String())
	if err != nil {
		return oapi.Up500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: fmt.Sprintf("failed to get env: %s", err)}}, nil
	}
	if env.TeamId != token.TeamId {
		return oapi.Up400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: "env does not belong to team"}}, nil
	}

	// Reset the file pointer to the beginning
	if _, err := tempFile.Seek(0, 0); err != nil {
		return oapi.Up500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: "failed to reset file pointer"}}, nil
	}

	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "archive-")
	if err != nil {
		return oapi.Up500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: "Failed to create temporary directory"}}, nil
	}
	fmt.Println("DEBUG: tempDir", tempDir)
	//defer os.RemoveAll(tempDir)

	// Untar and ungzip the archive
	gzr, err := gzip.NewReader(tempFile)
	if err != nil {
		return oapi.Up400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: "Failed to create gzip reader"}}, nil
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return oapi.Up400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: "Failed to read tar archive"}}, nil
		}

		target := filepath.Join(tempDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return oapi.Up500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: "Failed to create directory"}}, nil
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return oapi.Up500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: "Failed to create file"}}, nil
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return oapi.Up500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: "Failed to write file"}}, nil
			}
			f.Close()
		}
	}

	build, err := a.buildStore.Init(ctx, store.InitBuildOptions{
		TeamId:    token.TeamId,
		CreatorId: token.CreatorId,
		AppId:     app.Id,
	})
	if err != nil {
		return oapi.Up500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: fmt.Sprintf("failed to initialize build: %s", err)}}, nil
	}

	return customUpResponse{
		ctx:                 ctx,
		buildStore:          a.buildStore,
		cellStore:           a.cellStore,
		cellProviderForType: a.cellProviderForType,
		build:               build,
		tempDir:             tempDir,
		app:                 app,
		env:                 env,
		token:               token,
	}, nil
}

type customUpResponse struct {
	ctx                 context.Context
	buildStore          store.BuildStore
	cellStore           store.CellStore
	cellProviderForType func(cellType store.CellType) cellprovider.CellProvider
	build               store.Build
	tempDir             string
	app                 store.App
	env                 store.Env
	token               store.ApiToken
}

type flusherWriter struct {
	w http.ResponseWriter
}

func (fw *flusherWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if err == nil {
		if flusher, ok := fw.w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
	return n, err
}

func (c customUpResponse) VisitUpResponse(w http.ResponseWriter) (err error) {
	defer func() {
		if err != nil {
			c.buildStore.UpdateStatus(context.Background(), c.build.Id, store.BuildStatusFailed, err.Error())
		} else {
			c.buildStore.UpdateStatus(context.Background(), c.build.Id, store.BuildStatusCompleted, "")
		}
	}()
	logger := logger.FromContext(c.ctx).With("buildId", c.build.Id, "appId", c.app.Id, "appName", c.app.Name, "envId", c.env.Id, "envName", c.env.Name, "teamId", c.token.TeamId)
	defer os.Remove(c.tempDir)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	err = c.buildStore.UpdateStatus(context.Background(), c.build.Id, store.BuildStatusBuilding, "")
	if err != nil {
		fmt.Fprintf(w, "data: Error starting build: %s\n\n", err)
		return
	}

	logger.Info("build started")
	var cells []store.Cell
	cells, err = c.cellStore.GetForTeam(c.ctx, c.token.TeamId)
	if err != nil {
		logger.Error("failed to get cells", "error", err)
		return
	}
	if len(cells) != 1 {
		logger.Error("TODO: support specifying selecting a subset of multiple cells for build+deploy", "numCells", len(cells))
		err = fmt.Errorf("TODO: %d cells. support specifying selecting a subset of multiple cells for build+deploy", len(cells))
		return
	}
	cell := cells[0]

	cp := c.cellProviderForType(cell.Type)
	var artifact *store.ImageArtifact
	fw := &flusherWriter{w: w}
	artifact, err = cp.BuildImage(c.ctx, cellprovider.BuildImageOptions{
		CellId:   cell.Id,
		BuildDir: c.tempDir,
		AppName:  c.app.Name,
		BuildId:  c.build.Id,
		Stdout:   fw,
		Stderr:   fw,
	})
	if err != nil {
		err = fmt.Errorf("failed to build image: %w", err)
		return
	}
	w.(http.Flusher).Flush()

	if err := c.buildStore.UpdateArtifacts(c.ctx, c.build.Id, []store.BuildArtifact{
		{Image: artifact},
	}); err != nil {
		return fmt.Errorf("failed to update build artifacts: %w", err)
	}
	fmt.Fprintf(w, "build complete\n")
	w.(http.Flusher).Flush()

	// TODO: deploy the build

	return nil
}
