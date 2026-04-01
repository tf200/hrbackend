package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	modeCheck = "check"
	modeSync  = "sync"
)

var permissionPattern = regexp.MustCompile(`requirePermission\("([A-Z0-9_.]+)"\)`)

func main() {
	mode := modeCheck
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	if mode != modeCheck && mode != modeSync {
		fmt.Fprintf(os.Stderr, "Usage: go run ./scripts/permissions_catalog [check|sync]\n")
		os.Exit(1)
	}

	repoRoot, err := os.Getwd()
	if err != nil {
		fail("get working directory", err)
	}

	handlerDir := filepath.Join(repoRoot, "internal", "handler")
	catalogPath := filepath.Join(repoRoot, "migrations", "permissions_catalog.txt")

	actual, err := extractPermissions(handlerDir)
	if err != nil {
		fail("extract permissions", err)
	}

	expected, err := readCatalog(catalogPath)
	if err != nil {
		fail("read catalog", err)
	}

	switch mode {
	case modeSync:
		if err := writeCatalog(catalogPath, actual); err != nil {
			fail("write catalog", err)
		}
		fmt.Printf("Permission catalog synced: %s\n", catalogPath)
	case modeCheck:
		missing := difference(actual, expected)
		stale := difference(expected, actual)
		if len(missing) > 0 || len(stale) > 0 {
			fmt.Println("Permission catalog drift detected.")
			if len(missing) > 0 {
				fmt.Println()
				fmt.Println("Present in handlers, missing in catalog:")
				for _, p := range missing {
					fmt.Println(p)
				}
			}
			if len(stale) > 0 {
				fmt.Println()
				fmt.Println("Present in catalog, missing in handlers:")
				for _, p := range stale {
					fmt.Println(p)
				}
			}
			fmt.Println()
			fmt.Println("Run: go run ./scripts/permissions_catalog sync")
			os.Exit(1)
		}
		fmt.Println("Permission catalog is up to date.")
	}
}

func fail(step string, err error) {
	fmt.Fprintf(os.Stderr, "permissions-catalog: %s: %v\n", step, err)
	os.Exit(1)
}

func extractPermissions(handlerDir string) ([]string, error) {
	seen := make(map[string]struct{})

	err := filepath.WalkDir(handlerDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		matches := permissionPattern.FindAllSubmatch(content, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			seen[string(match[1])] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(seen) == 0 {
		return nil, errors.New("no permissions found under internal/handler")
	}

	out := make([]string, 0, len(seen))
	for permission := range seen {
		out = append(out, permission)
	}
	sort.Strings(out)
	return out, nil
}

func readCatalog(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		seen[line] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	out := make([]string, 0, len(seen))
	for permission := range seen {
		out = append(out, permission)
	}
	sort.Strings(out)
	return out, nil
}

func writeCatalog(path string, permissions []string) error {
	var buf bytes.Buffer
	for _, permission := range permissions {
		buf.WriteString(permission)
		buf.WriteByte('\n')
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func difference(left, right []string) []string {
	inRight := make(map[string]struct{}, len(right))
	for _, item := range right {
		inRight[item] = struct{}{}
	}

	out := make([]string, 0)
	for _, item := range left {
		if _, ok := inRight[item]; !ok {
			out = append(out, item)
		}
	}
	return out
}
