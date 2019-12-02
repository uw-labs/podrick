package podman

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/uw-labs/podrick"
	podman "github.com/uw-labs/podrick/runtimes/podman/iopodman"
	"github.com/varlink/go/varlink"
)

func uploadFiles(ctx context.Context, conn *varlink.Connection, cID string, files ...podrick.File) (err error) {
	mountDir, err := podman.MountContainer().Call(ctx, conn, cID)
	if err != nil {
		return fmt.Errorf("failed to mount container filesystem: %w", err)
	}
	defer func() {
		uErr := podman.UnmountContainer().Call(context.Background(), conn, cID, true)
		if err == nil {
			err = uErr
		}
	}()

	for _, f := range files {
		err = uploadFile(ctx, mountDir, f)
		if err != nil {
			return fmt.Errorf("failed to upload file: %w", err)
		}
	}

	return nil
}

func uploadFile(ctx context.Context, mountDir string, file podrick.File) (err error) {
	path := filepath.Clean(file.Path)
	if !filepath.IsAbs(path) {
		return fmt.Errorf("file paths must be absolute: %q", file.Path)
	}
	dest := filepath.Join(mountDir, path)
	if _, err := os.Stat(filepath.Dir(dest)); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(filepath.Dir(dest), 0777)
		if err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
	}
	target, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		cErr := target.Close()
		if err == nil {
			err = cErr
		}
	}()
	_, err = io.Copy(target, file.Content)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
