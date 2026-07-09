package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"

	ignore "github.com/sabhiram/go-gitignore"
)

// findGitRoot traverses from the current directory all the way to parents
// and then finds the git root by looking for a folder called ".git" - if
// it finds it - it returns it otherwise, simply returns an error if it can't
// find one.
func findGitRoot(dir string) (string, error) {
	base, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("unable to find a valid basepath: %v", err)
	}

	for {
		fileInfo, err := os.Stat(filepath.Join(base, ".git"))
		if err == nil && fileInfo.IsDir() {
			return base, nil
		}
		parent := filepath.Dir(base)
		if parent == base {
			return "", fmt.Errorf("unable to find git root, reached /")
		}
		base = parent
	}
}

type searchResult struct {
	lineNo      int
	filename    string
	lineSnippet string
}

func searchFile(filepath string, pattern *regexp.Regexp, gi *ignore.GitIgnore) []searchResult {
	if gi.MatchesPath(filepath) {
		return nil
	}
	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("error reading file : %s due to %v - skipping", filepath, err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	var searchResults []searchResult
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber += 1
		matches := pattern.FindAllString(line, -1)
		if len(matches) > 0 {
			searchResults = append(searchResults, searchResult{
				lineNo:      lineNumber,
				filename:    filepath,
				lineSnippet: matches[0],
			})
		}
	}

	return searchResults
}

func getWalkFunc(pattern *regexp.Regexp, gi *ignore.GitIgnore) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if gi.MatchesPath(path) {
			return filepath.SkipDir
		}
		for _, searchResult := range searchFile(path, pattern, gi) {
			fmt.Printf("%s:%d:0: %s\n", searchResult.filename, searchResult.lineNo, searchResult.lineSnippet)
		}

		return nil
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("required atleast one argument - the pattern")
	}
	pattern := regexp.MustCompile(os.Args[1])
	baseDir := "."
	if len(os.Args) == 3 {
		baseDir = os.Args[2]
	}
	gitRoot, err := findGitRoot(baseDir)
	if err != nil {
		gitRoot = "."
	}
	_, err = os.Stat(filepath.Join(gitRoot, ".gitignore"))
	gi := ignore.CompileIgnoreLines([]string{"node_modules", "*.pyc", "~"}...)
	if err != nil {
		// we might not have an ignore file, so we just log a warning and
		// move on.
		log.Printf("unable to find .gitignore in %s due to %v", gitRoot, err)
	} else {
		gi2, err := ignore.CompileIgnoreFile(filepath.Join(gitRoot, ".gitignore"))
		if err == nil {
			gi = gi2
		}
	}

	filepath.Walk(gitRoot, getWalkFunc(pattern, gi))
}
