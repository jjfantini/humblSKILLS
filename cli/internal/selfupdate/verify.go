package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// downloadToFile streams url into a tmp file beside dest, then renames it
// into place atomically — same convention as internal/fetch.Fetcher.Fetch
// and internal/registry.writeAtomic.
func downloadToFile(client *http.Client, url, dest string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fetch %s: HTTP %d", url, resp.StatusCode)
	}

	tmp := dest + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create tmp file: %w", err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("write %s: %w", dest, err)
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("finalise %s: %w", dest, err)
	}
	return nil
}

// downloadChecksums fetches a checksums.txt-style manifest and returns a map
// of asset name -> lowercase hex sha256.
func downloadChecksums(client *http.Client, url string) (map[string]string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch %s: HTTP %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read checksums: %w", err)
	}
	return parseChecksums(string(body)), nil
}

// parseChecksums parses sha256sum-style lines ("<hex>  <filename>", or the
// binary-mode "<hex> *<filename>" variant) into a name -> hex map. Uses
// strings.Fields so it doesn't care whether the separator is one or two
// spaces.
func parseChecksums(body string) map[string]string {
	sums := map[string]string{}
	for _, line := range strings.Split(body, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		hash := strings.ToLower(fields[0])
		name := strings.TrimPrefix(fields[len(fields)-1], "*")
		sums[name] = hash
	}
	return sums
}

// sha256File computes the lowercase hex sha256 of the file at path.
func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyChecksum confirms the file at archivePath matches the checksum
// published for assetName in sums.
func VerifyChecksum(archivePath, assetName string, sums map[string]string) error {
	want, ok := sums[assetName]
	if !ok {
		return fmt.Errorf("no checksum entry for %s", assetName)
	}
	got, err := sha256File(archivePath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(want, got) {
		return fmt.Errorf("checksum mismatch for %s: want %s got %s", assetName, want, got)
	}
	return nil
}
