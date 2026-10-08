package bootimg

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"
)

const fetchTimeout = 5 * time.Minute

func (i Image) Resolve(ctx context.Context, progress func(float64)) ([]byte, string, error) {
	if i.URL == "" {
		return nil, "", fmt.Errorf(
			"bootimg: no boot image is published for %s, so one has to be given with --boot-image",
			i.Device,
		)
	}

	data, err := i.download(ctx, progress)
	if err != nil {
		return nil, "", err
	}
	if err := i.Verify(i.URL, data); err != nil {
		return nil, "", err
	}
	return data, i.URL, nil
}

func (i Image) download(ctx context.Context, progress func(float64)) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, i.URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bootimg: fetching %s: %w", i.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bootimg: fetching %s: %s", i.URL, resp.Status)
	}

	buf := make([]byte, 0, i.Size)
	sum := sha256.New()
	body := io.LimitReader(resp.Body, i.Size+1)

	chunk := make([]byte, 64*1024)
	for {
		n, err := body.Read(chunk)
		if n > 0 {
			buf = append(buf, chunk[:n]...)
			sum.Write(chunk[:n])
			if progress != nil && i.Size > 0 {
				progress(min(float64(len(buf))/float64(i.Size), 1))
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("bootimg: fetching %s: %w", i.URL, err)
		}
	}

	if got := hex.EncodeToString(sum.Sum(nil)); got != i.SHA256 {
		return nil, fmt.Errorf("bootimg: %s hashes to %s, want %s", i.URL, got, i.SHA256)
	}
	return buf, nil
}
