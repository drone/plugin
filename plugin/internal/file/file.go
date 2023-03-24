package file

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/drone/plugin/cache"
	"github.com/klauspost/compress/zstd"
	"github.com/pkg/errors"
)

const (
	defaultDownloadTimeout = 300 * time.Second
)

func Download(url string) (string, error) {
	key := cache.GetKeyName(url)
	binPath := filepath.Join(key, "step.exe")

	downloadFn := func() error {
		if err := download(url, binPath); err != nil {
			return errors.Wrap(err, fmt.Sprintf("url: %s", url))
		}
		return nil
	}

	if err := cache.Add(key, downloadFn); err != nil {
		return "", err
	}
	return binPath, nil
}

// download method downloads a source & writes it to a file.
// If file is compressed, it also decompresses the file on the basis of
// file extension in url. Currently it supports zstd format.
func download(url, path string) error {
	client := http.Client{
		Timeout: getDownloadTimeout(),
	}
	resp, err := client.Get(url)
	if err != nil {
		return errors.Wrap(err, "failed to download url")
	}
	defer resp.Body.Close()

	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create file at path: %s", path))
	}
	defer f.Close()

	if strings.HasSuffix(url, ".zst") {
		return decompress(resp.Body, f)
	}

	if _, err = io.Copy(f, resp.Body); err != nil {
		return errors.Wrap(err, "failed to write download binary to file")
	}

	return nil
}

func decompress(in io.Reader, out io.Writer) error {
	d, err := zstd.NewReader(in)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(out, d)
	return err
}

func getDownloadTimeout() time.Duration {
	timeout, err := strconv.Atoi(os.Getenv("DRONE_DOWNLOAD_TIMEOUT_SECS"))
	if err == nil && timeout > 0 {
		return time.Duration(timeout) * time.Second
	}
	return defaultDownloadTimeout
}
