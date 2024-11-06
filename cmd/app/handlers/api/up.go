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
	"time"

	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/lib/background"
	"github.com/onmetal-dev/metal/lib/background/deployment"
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
	defer os.RemoveAll(tempDir)

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
		deploymentStore:     a.deploymentStore,
		appStore:            a.appStore,
		producerDeployment:  a.producerDeployment,
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
	deploymentStore     store.DeploymentStore
	appStore            store.AppStore
	producerDeployment  *background.QueueProducer[deployment.Message]
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

	if err := c.buildStore.UpdateArtifacts(c.ctx, c.build.Id, []store.Artifact{
		{Image: artifact},
	}); err != nil {
		return fmt.Errorf("failed to update build artifacts: %w", err)
	}
	fmt.Fprintf(fw, "✅ build complete! beginning deployment 🚀\n")

	// get latest deployment for this app in this env. We will clone the app settings from this deployment so we match the previous deployment as much as possible.
	ld, err := c.deploymentStore.GetLatestForAppEnv(c.ctx, c.app.Id, c.env.Id)
	if err != nil {
		return fmt.Errorf("failed to get latest deployment: %w", err)
	}

	var appSettings *store.AppSettings
	var appEnvVars *store.AppEnvVars
	if ld != nil {
		appEnvVars = &ld.AppEnvVars
		as, err := c.appStore.CreateAppSettings(store.CreateAppSettingsOptions{
			TeamId: c.token.TeamId,
			AppId:  c.app.Id,
			Artifact: store.Artifact{
				Image: artifact,
			},
			Ports:         ld.AppSettings.Ports.Data(),
			ExternalPorts: ld.AppSettings.ExternalPorts.Data(),
			Resources:     ld.AppSettings.Resources.Data(),
		})
		if err != nil {
			return fmt.Errorf("failed to create app settings: %w", err)
		}
		appSettings = &as
	} else {
		as, err := c.appStore.CreateAppSettings(store.CreateAppSettingsOptions{
			TeamId: c.token.TeamId,
			AppId:  c.app.Id,
			Artifact: store.Artifact{
				Image: artifact,
			},
			Ports: []store.Port{
				{
					Name:  "http",
					Port:  80,
					Proto: "http",
				},
			},
			ExternalPorts: []store.ExternalPort{
				{
					Name:  "http",
					Port:  80,
					Proto: "http",
				},
			},
			Resources: store.Resources{
				Limits: store.ResourceLimits{
					CpuCores:  0.1,
					MemoryMiB: 128,
				},
				Requests: store.ResourceRequests{
					CpuCores:  0.1,
					MemoryMiB: 128,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create app settings: %w", err)
		}
		appSettings = &as

		aev, err := c.deploymentStore.CreateAppEnvVars(store.CreateAppEnvVarOptions{
			TeamId:  c.token.TeamId,
			EnvId:   c.env.Id,
			AppId:   c.app.Id,
			EnvVars: []store.EnvVar{},
		})
		if err != nil {
			return fmt.Errorf("failed to create app env vars: %w", err)
		}
		appEnvVars = &aev
	}

	cdo := store.CreateDeploymentOptions{
		TeamId:        c.token.TeamId,
		EnvId:         c.env.Id,
		AppId:         c.app.Id,
		Type:          store.DeploymentTypeDeploy,
		AppSettingsId: appSettings.Id,
		AppEnvVarsId:  appEnvVars.Id,
		CellIds:       []string{cell.Id},
		Replicas:      1,
	}
	if ld != nil {
		cdo.Replicas = ld.Replicas
		cdo.CellIds = lo.Map(ld.Cells, func(cell store.Cell, _ int) string {
			return cell.Id
		})
	}
	d, err := c.deploymentStore.Create(cdo)
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}
	if err := c.producerDeployment.Send(c.ctx, deployment.Message{
		DeploymentId: d.Id,
		AppId:        d.AppId,
		EnvId:        d.EnvId,
	}); err != nil {
		return fmt.Errorf("failed to send deployment message to queue: %w", err)
	}

	errChan := make(chan error, 2)
	doneChan := make(chan struct{})

	// goroutine to poll deployment status
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.ctx.Done():
				errChan <- c.ctx.Err()
				return
			case <-doneChan:
				return
			case <-ticker.C:
				updatedD, err := c.deploymentStore.Get(d.AppId, d.EnvId, d.Id)
				if err != nil {
					errChan <- fmt.Errorf("failed to get deployment: %w", err)
					return
				}
				switch updatedD.Status {
				case store.DeploymentStatusRunning:
					close(doneChan)
					return
				case store.DeploymentStatusFailed:
					errChan <- fmt.Errorf("deployment failed: %s", updatedD.StatusReason)
					return
				}
			}
		}
	}()

	// goroutine to stream logs
	go func() {
		logChan := cp.DeploymentLogsStream(c.ctx, cell.Id, &d, cellprovider.WithSince(time.Minute*10))
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-doneChan:
				return
			case log, ok := <-logChan:
				if !ok {
					return
				}
				if log.Error != nil {
					errChan <- fmt.Errorf("failed to stream logs: %w", log.Error)
					return
				}
				for _, log := range log.Logs {
					if _, err := fmt.Fprintf(fw, "%s %s", log.Timestamp.Format(time.RFC3339), log.Message); err != nil {
						errChan <- fmt.Errorf("failed to write log: %w", err)
						return
					}
				}
			}
		}
	}()

	// wait for completion or error
	select {
	case err := <-errChan:
		return err
	case <-doneChan:
		return nil
	}
}
