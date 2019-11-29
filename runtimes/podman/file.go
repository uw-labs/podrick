package podman

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/uw-labs/podrick"
	podman "github.com/uw-labs/podrick/runtimes/podman/iopodman"
	"github.com/varlink/go/varlink"
)

func uploadFiles(ctx context.Context, addr string, files ...podrick.File) (err error) {
	for _, f := range files {
		uploadFile(ctx, addr, f)
	}

	return nil
}

func uploadFile(ctx context.Context, addr string, file podrick.File) (err error) {
	path := filepath.Clean(file.Path)
	if !filepath.IsAbs(path) {
		return fmt.Errorf("file paths must be absolute: %q", file.Path)
	}
	//dest := filepath.Join(mountDir, path)

	fConn, err := varlink.NewConnection(ctx, addr)
	if err != nil {
		return fmt.Errorf("failed to create new connection: %w", err)
	}
	defer func() {
		cErr := fConn.Close()
		if err == nil {
			err = cErr
		}
	}()

	reply, err := podman.SendFile().Upgrade(ctx, fConn, "", int64(file.Size))
	if err != nil {
		return fmt.Errorf("failed to start connection upgrade: %w", err)
	}

	_, _, conn, err := reply(ctx)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %w", err)
	}

	w := writerCtx{
		WriterContext: conn,
		ctx:           ctx,
	}

	_, err = io.Copy(w, file.Content)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	filename, err := conn.ReadBytes(ctx, ':')
	if err != nil {
		return fmt.Errorf("failed to read file name: %w", err)
	}

	podman.LoadImage().Call(ctx, fConn)

	fmt.Println("Create file", strings.ReplaceAll(string(filename), ":", ""))

	return nil
}

type writerCtx struct {
	varlink.WriterContext
	ctx context.Context
}

func (w writerCtx) Write(in []byte) (int, error) {
	return w.WriterContext.Write(w.ctx, in)
}
