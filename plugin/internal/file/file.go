package file

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/pkg/errors"
)

// Download method downloads a source & writes it to a file.
// If file is compressed, it also decompresses the file on the basis of
// file extension in url. Currently it supports zstd format.
func Download(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to download url: %s", url))
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
