package external

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bananenpro/cli"
	"github.com/adrg/xdg"

	"github.com/code-game-project/go-utils/semver"
)

var (
	ErrTagNotFound = errors.New("tag not found")
	ErrHTTPStatus  = errors.New("invalid http status")
)

func LatestGithubTag(owner, repo string) (string, error) {
	type response []struct {
		Name string `json:"name"`
	}
	res, err := githubRequest[response](fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo))
	if err != nil {
		return "", fmt.Errorf("failed to access git tags from 'github.com/%s/%s'.", owner, repo)
	}
	return res[0].Name, nil
}

func GithubTagFromVersion(owner, repo, version string) (string, error) {
	type response []struct {
		Name string `json:"name"`
	}
	res, err := githubRequest[response](fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo))
	if err != nil {
		return "", fmt.Errorf("Couldn't access git tags from 'github.com/%s/%s'.", owner, repo)
	}

	for _, tag := range res {
		if strings.HasPrefix(tag.Name, "v"+version) {
			return tag.Name, nil
		}
	}
	return "", ErrTagNotFound
}

func LibraryVersionFromCGVersion(owner, repo, cgVersion string) string {
	res, err := LoadVersionsJSON(owner, repo)
	if err != nil {
		cli.Warn("Couldn't fetch versions.json. Using latest client library version.")
		return "latest"
	}

	var versions map[string]string
	err = json.Unmarshal(res, &versions)
	if err != nil {
		cli.Warn("Invalid versions.json. Using latest client library version.")
		return "latest"
	}

	return semver.CompatibleVersion(versions, cgVersion)
}

func LoadVersionsJSON(owner, repo string) (json.RawMessage, error) {
	res, err := githubRequest[json.RawMessage](fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/versions.json", owner, repo))
	if err != nil {
		res, err = githubRequest[json.RawMessage](fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/versions.json", owner, repo))
	}
	return res, err
}

func githubRequest[T any](url string) (T, error) {
	var data T

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return data, err
	}

	etag, err := githubETag(url)
	if err == nil {
		request.Header.Add("If-None-Match", etag)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err == nil {
			saveGitHubETag(url, resp.Header.Get("etag"), data)
		}
		return data, err
	}
	if resp.StatusCode == http.StatusNotModified {
		return loadGitHubCacheData[T](url)
	}

	return data, ErrHTTPStatus
}

var githubCacheDir = filepath.Join(xdg.CacheHome, "codegame", "github_requests")

func saveGitHubETag(reqURL, etag string, data any) error {
	os.MkdirAll(githubCacheDir, 0o755)

	file, err := os.Create(filepath.Join(githubCacheDir, url.PathEscape(reqURL)))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(etag + "\n")
	if err != nil {
		return err
	}
	return json.NewEncoder(file).Encode(data)
}

func githubETag(reqURL string) (string, error) {
	file, err := os.Open(filepath.Join(githubCacheDir, url.PathEscape(reqURL)))
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if scanner.Err() != nil {
			return "", scanner.Err()
		} else {
			return "", io.EOF
		}
	}
	return scanner.Text(), nil
}

func loadGitHubCacheData[T any](reqURL string) (T, error) {
	var data T

	file, err := os.Open(filepath.Join(githubCacheDir, url.PathEscape(reqURL)))
	if err != nil {
		return data, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if scanner.Err() != nil {
			return data, scanner.Err()
		} else {
			return data, io.EOF
		}
	}
	if !scanner.Scan() {
		if scanner.Err() != nil {
			return data, scanner.Err()
		} else {
			return data, io.EOF
		}
	}

	err = json.Unmarshal(scanner.Bytes(), &data)
	return data, err
}
