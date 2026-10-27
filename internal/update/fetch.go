package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	httpfetch "github.com/ygelfand/libcountertop/pkg/fetch"
)

func At(ctx context.Context, src string) (Manifest, error) {
	var m Manifest

	if path, ok := local(src); ok {
		data, err := os.ReadFile(path)
		if err != nil {
			return m, err
		}
		if err := json.Unmarshal(data, &m); err != nil {
			return m, fmt.Errorf("update: reading the manifest at %s: %w", path, err)
		}
		return m, m.Valid()
	}

	ctx, cancel := context.WithTimeout(ctx, manifestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return m, err
	}

	resp, err := httpfetch.Client(0).Do(req)
	if err != nil {
		return m, fmt.Errorf("update: fetching %s: %w", src, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return m, fmt.Errorf("update: fetching %s: %s", src, resp.Status)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxManifest)).Decode(&m); err != nil {
		return m, fmt.Errorf("update: reading the manifest at %s: %w", src, err)
	}
	return m, m.Valid()
}

func local(src string) (string, bool) {
	if path, ok := strings.CutPrefix(src, "file://"); ok {
		return path, true
	}
	if strings.Contains(src, "://") {
		return "", false
	}
	return src, true
}

const binaryTimeout = 5 * time.Minute

func (b Binary) Resolve(ctx context.Context, progress func(float64)) ([]byte, string, error) {
	if path, ok := local(b.URL); ok {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, "", err
		}
		if err := b.check(b.URL, data); err != nil {
			return nil, "", err
		}
		if progress != nil {
			progress(1)
		}
		return data, b.URL, nil
	}

	ctx, cancel := context.WithTimeout(ctx, binaryTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.URL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := httpfetch.Client(0).Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("update: fetching %s: %w", b.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("update: fetching %s: %s", b.URL, resp.Status)
	}

	buf := make([]byte, 0, b.Size)
	body := io.LimitReader(resp.Body, b.Size+1)

	chunk := make([]byte, 64*1024)
	for {
		n, err := body.Read(chunk)
		if n > 0 {
			buf = append(buf, chunk[:n]...)
			if progress != nil && b.Size > 0 {
				progress(min(float64(len(buf))/float64(b.Size), 1))
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("update: fetching %s: %w", b.URL, err)
		}
	}

	if err := b.check(b.URL, buf); err != nil {
		return nil, "", err
	}
	return buf, b.URL, nil
}

func (b Binary) check(from string, data []byte) error {
	if int64(len(data)) != b.Size {
		return fmt.Errorf("update: %s is %d bytes, want %d", from, len(data), b.Size)
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != b.SHA256 {
		return fmt.Errorf("update: %s hashes to %s, want %s", from, got, b.SHA256)
	}
	return nil
}
