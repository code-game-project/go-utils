package cggenevents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"

	"github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/external"
)

var cgGenEventsPath = filepath.Join(xdg.DataHome, "codegame", "bin", "cg-gen-events")

// LatestCGEVersion returns the latest CGE version in the format 'x.y'.
func LatestCGEVersion() (string, error) {
	tag, err := external.LatestGithubTag("code-game-project", "cg-gen-events")
	if err != nil {
		return "", fmt.Errorf("Couldn't determine the latest CGE version: %s", err)
	}

	return strings.TrimPrefix(strings.Join(strings.Split(tag, ".")[:2], "."), "v"), nil
}

// CGGenEvents downloads and executes the correct version of cg-gen-events.
// cgePath can be either a filepath or a url starting with either http:// or https://.
func CGGenEvents(cgeVersion, outputDir, cgePath, languages string) error {
	exeName, err := InstallCGGenEvents(cgeVersion)
	if err != nil {
		return err
	}
	_, err = exec.Execute(true, filepath.Join(cgGenEventsPath, exeName), cgePath, "-l", languages, "-o", outputDir)
	return err
}

// InstallCGGenEvents installs the correct version of cg-gen-events if neccessary.
func InstallCGGenEvents(cgeVersion string) (string, error) {
	version, err := external.GithubTagFromVersion("code-game-project", "cg-gen-events", cgeVersion)
	if err != nil {
		return "", err
	}
	version = strings.TrimPrefix(version, "v")
	return external.InstallProgram("cg-gen-events", "cg-gen-events", fmt.Sprintf("https://github.com/code-game-project/cg-gen-events"), version, cgGenEventsPath)
}

// GetEventNames uses CGGenEvents() to get a list of all the available event and command names of the game server at url.
// It only works for CGE versions >= 0.3.
func GetEventNames(url, cgeVersion string) (eventNames []string, commandNames []string, err error) {
	output := os.TempDir()
	err = CGGenEvents(cgeVersion, output, url, "json")
	if err != nil {
		return nil, nil, err
	}

	type obj struct {
		Name string `json:"name"`
	}
	type data struct {
		Events   []obj `json:"events"`
		Commands []obj `json:"commands"`
	}

	path := filepath.Join(output, "events.json")

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer os.Remove(path)
	defer file.Close()

	var object data
	err = json.NewDecoder(file).Decode(&object)
	if err != nil {
		return nil, nil, err
	}

	eNames := make([]string, len(object.Events))
	for i, event := range object.Events {
		eNames[i] = event.Name
	}

	cNames := make([]string, len(object.Commands))
	for i, cmd := range object.Commands {
		cNames[i] = cmd.Name
	}
	return eNames, cNames, nil
}

// CGEVersion parses the version field in the provided CGE file and returns the CGE version in the format 'x.y'.
func ParseCGEVersion(cge string) (string, error) {
	runes := []rune(cge)
	index := 0
	commentNestingLevel := 0
	for index < len(runes) && (runes[index] == ' ' || runes[index] == '\r' || runes[index] == '\n' || runes[index] == '\t' || (index < len(runes)-1 && runes[index] == '/' && runes[index+1] == '*') || (index < len(runes)-1 && runes[index] == '*' && runes[index+1] == '/') || (index < len(runes)-1 && runes[index] == '/' && runes[index+1] == '/') || commentNestingLevel > 0) {
		if runes[index] == '/' {
			if runes[index+1] == '/' {
				for index < len(runes) && runes[index] != '\n' {
					index++
				}
			} else {
				commentNestingLevel++
			}
		}
		if runes[index] == '*' {
			commentNestingLevel--
		}
		index++
	}

	words := strings.Fields(string(runes[index:]))
	for i, w := range words {
		if w == "version" && i < len(words)-1 {
			return words[i+1], nil
		}
	}

	return "", fmt.Errorf("invalid CGE file: no version field")
}
