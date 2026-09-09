// Package github fetches public repository metadata only. It never executes repository code.
package github

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	HTTP     *http.Client
	MaxBytes int64
}
type Files struct {
	CommitSHA             string
	Pyproject, PoetryLock []byte
}

var ownerPattern = regexp.MustCompile("^[A-Za-z0-9-]+$")
var repoPattern = regexp.MustCompile("^[A-Za-z0-9_.-]+$")
var shaPattern = regexp.MustCompile("^[a-fA-F0-9]{40}$")

func ParseRepositoryURL(raw string) (string, string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", "", fmt.Errorf("INVALID_REPOSITORY_URL: expected public https://github.com/owner/repository URL")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("INVALID_REPOSITORY_URL: expected owner/repository path")
	}
	owner, repo := parts[0], strings.TrimSuffix(parts[1], ".git")
	if !ownerPattern.MatchString(owner) || !repoPattern.MatchString(repo) || repo == "." || repo == ".." {
		return "", "", fmt.Errorf("INVALID_REPOSITORY_URL: invalid owner or repository")
	}
	return owner, repo, nil
}
func (c Client) Fetch(ctx context.Context, raw string) (Files, error) {
	owner, repo, err := ParseRepositoryURL(raw)
	if err != nil {
		return Files{}, err
	}
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/HEAD", owner, repo), nil)
	if err != nil {
		return Files{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	response, err := client.Do(req)
	if err != nil {
		return Files{}, fmt.Errorf("GITHUB_UNAVAILABLE: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return Files{}, fmt.Errorf("GITHUB_METADATA_ERROR: GitHub returned %s", response.Status)
	}
	var commit struct{ SHA string }
	if err = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&commit); err != nil || !shaPattern.MatchString(commit.SHA) {
		return Files{}, fmt.Errorf("GITHUB_METADATA_ERROR: invalid commit response")
	}
	req, err = http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://codeload.github.com/%s/%s/tar.gz/%s", owner, repo, commit.SHA), nil)
	if err != nil {
		return Files{}, err
	}
	response, err = client.Do(req)
	if err != nil {
		return Files{}, fmt.Errorf("GITHUB_DOWNLOAD_ERROR: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return Files{}, fmt.Errorf("GITHUB_DOWNLOAD_ERROR: GitHub returned %s", response.Status)
	}
	limit := c.MaxBytes
	if limit <= 0 {
		limit = 10 << 20
	}
	archive, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return Files{}, fmt.Errorf("GITHUB_DOWNLOAD_ERROR: %w", err)
	}
	if int64(len(archive)) > limit {
		return Files{}, fmt.Errorf("REPOSITORY_FILE_TOO_LARGE: archive exceeds download limit")
	}
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return Files{}, fmt.Errorf("GITHUB_DOWNLOAD_ERROR: invalid archive: %w", err)
	}
	defer gz.Close()
	expanded := &io.LimitedReader{R: gz, N: 50 << 20}
	tr := tar.NewReader(expanded)
	files := Files{CommitSHA: commit.SHA}
	for {
		header, err := tr.Next()
		if expanded.N == 0 {
			return Files{}, fmt.Errorf("REPOSITORY_FILE_TOO_LARGE: expanded archive exceeds 50 MiB")
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return Files{}, fmt.Errorf("GITHUB_DOWNLOAD_ERROR: invalid archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		name := path.Clean(header.Name)
		if strings.Count(name, "/") != 1 || strings.HasPrefix(name, "/") || (path.Base(name) != "pyproject.toml" && path.Base(name) != "poetry.lock") {
			continue
		}
		if header.Size > 1<<20 {
			return Files{}, fmt.Errorf("REPOSITORY_FILE_TOO_LARGE: %s exceeds 1 MiB", name)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return Files{}, fmt.Errorf("GITHUB_DOWNLOAD_ERROR: %w", err)
		}
		switch path.Base(name) {
		case "pyproject.toml":
			files.Pyproject = data
		case "poetry.lock":
			files.PoetryLock = data
		}
	}
	if len(files.Pyproject) == 0 || len(files.PoetryLock) == 0 {
		return Files{}, fmt.Errorf("UNSUPPORTED_INPUT: paired root pyproject.toml and poetry.lock are required; requirements.txt-only and manifest-only scans are deferred")
	}
	return files, nil
}
