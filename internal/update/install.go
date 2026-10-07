package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	httpfetch "github.com/ygelfand/libcountertop/pkg/fetch"

	"github.com/ygelfand/echolocal/internal/layout"
)

const downloadTimeout = 10 * time.Minute

var installing sync.Mutex

func Install(ctx context.Context, m Manifest, progress func(float32)) error {
	if !installing.TryLock() {
		return fmt.Errorf("update: an install is already running")
	}
	defer installing.Unlock()

	if err := refusePryon(); err != nil {
		return err
	}
	if err := m.Valid(); err != nil {
		return err
	}
	b, err := m.For(codename, arch)
	if err != nil {
		return err
	}

	staged := filepath.Join(layout.StateDir, "echod.incoming")
	defer os.Remove(staged)

	if err := download(ctx, b, staged, progress); err != nil {
		return err
	}
	tidy()
	if err := room(b.Size); err != nil {
		return err
	}
	return swap(staged, m.Version)
}

func download(ctx context.Context, b Binary, to string, progress func(float32)) error {
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.URL, nil)
	if err != nil {
		return err
	}
	resp, err := httpfetch.Client(0).Do(req)
	if err != nil {
		return fmt.Errorf("update: fetching %s: %w", b.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update: fetching %s: %s", b.URL, resp.Status)
	}

	f, err := os.OpenFile(to, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer f.Close()

	sum := sha256.New()
	written, err := io.Copy(io.MultiWriter(f, sum), &counter{
		from: resp.Body, size: b.Size, report: progress,
	})
	if err != nil {
		return fmt.Errorf("update: downloading %s: %d of %d bytes: %w", b.URL, written, b.Size, err)
	}
	if err := f.Sync(); err != nil {
		return err
	}

	if written != b.Size {
		return fmt.Errorf("update: %d bytes, offered as %d", written, b.Size)
	}
	if got := hex.EncodeToString(sum.Sum(nil)); got != b.SHA256 {
		return fmt.Errorf("update: hash %s, offered as %s", got, b.SHA256)
	}
	if progress != nil {
		progress(1)
	}
	return nil
}

func swap(staged, version string) error {
	if err := os.WriteFile(layout.UpdatingPath, []byte(version), 0o644); err != nil {
		slog.Error("recording the version being installed failed", "err", err)
	}

	part := prev + ".part"
	if err := copyTo(binary, part); err != nil {
		_ = os.Remove(part)
		return fmt.Errorf("update: keeping the running binary: %w", err)
	}
	if err := os.Rename(part, prev); err != nil {
		_ = os.Remove(part)
		return fmt.Errorf("update: keeping the running binary: %w", err)
	}

	if err := onSystem(func() error { return replace(staged) }); err != nil {
		if exists(binary) {
			_ = os.Remove(prev)
		}
		return fmt.Errorf("update: installing: %w", err)
	}
	slog.Info("update installed, restarting into it", "version", version, "previous", prev)
	return nil
}

func replace(staged string) error {
	if err := os.Rename(binary, aside); err != nil {
		return err
	}
	if err := copyTo(staged, binary); err != nil {
		_ = os.Remove(binary)
		if back := os.Rename(aside, binary); back != nil {
			slog.Error("could not put the running binary back", "err", back)
		}
		return err
	}

	if err := copyLabel(aside, binary); err != nil {
		slog.Warn("labelling the new binary failed", "err", err)
	}
	// ext4 refuses a read-only remount while a deleted file is still open.
	return nil
}

// /data and /system are different filesystems.
func copyTo(from, to string) error {
	src, err := os.Open(from)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(to, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("update: writing %s: %w", to, err)
	}
	return dst.Sync()
}

type counter struct {
	from   io.Reader
	size   int64
	report func(float32)

	read int64
	last float32
}

func (c *counter) Read(p []byte) (int, error) {
	n, err := c.from.Read(p)
	c.read += int64(n)

	if c.report == nil || c.size <= 0 {
		return n, err
	}

	if at := float32(c.read) / float32(c.size); at-c.last >= 0.01 {
		c.last = at
		c.report(at)
	}
	return n, err
}
