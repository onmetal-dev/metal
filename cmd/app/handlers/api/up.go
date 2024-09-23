package api

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/lib/oapi"
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
	//defer os.Remove(tempFile.Name())
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

	// TODO: Process the extracted files in tempDir

	return oapi.Up200JSONResponse{
		Message: lo.ToPtr("Archive successfully uploaded and extracted"),
		BuildId: nil, // TODO: Generate and return a build ID
	}, nil
}
