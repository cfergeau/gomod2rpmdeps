package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

type modScanner interface {
	Scan() bool
	Text() string
	Err() error
	io.Closer
}

type govendorScanner struct {
	*exec.Cmd
	scanner *bufio.Scanner
	err     error
}

func govendorScannerNew() *govendorScanner {
	cmd := exec.Command("go", "mod", "vendor", "-v")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return &govendorScanner{err: err}
	}
	if err := cmd.Start(); err != nil {
		return &govendorScanner{err: err}
	}
	scanner := bufio.NewScanner(stderr)

	return &govendorScanner{cmd, scanner, nil}
}

func (govendor *govendorScanner) Scan() bool {
	if govendor.err != nil {
		return false
	}

	return govendor.scanner.Scan()
}

func (govendor *govendorScanner) Text() string {
	if govendor.err != nil {
		return ""
	}

	return govendor.scanner.Text()
}

func (govendor *govendorScanner) Close() error {
	if govendor.err != nil {
		return govendor.err
	}

	return govendor.Wait()
}

func (govendor *govendorScanner) Err() error {
	if govendor.err != nil {
		return govendor.err
	}

	return govendor.scanner.Err()
}

type goModuleInfo struct {
	name          string
	pseudoVersion string
}

type byModuleName []goModuleInfo

func (a byModuleName) Len() int           { return len(a) }
func (a byModuleName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byModuleName) Less(i, j int) bool { return a[i].name < a[j].name }

var errNoSharpPrefix = fmt.Errorf("Parse error - line not starting with '#'")
var errComment = fmt.Errorf("Parse error - line is a comment")
var errEmptyLine = fmt.Errorf("Parse error - line is empty")

func parseModuleLine(line string) (goModuleInfo, error) {
	substrings := strings.Split(line, " ")
	if len(substrings) == 0 {
		return goModuleInfo{}, errEmptyLine
	}
	if substrings[0] != "#" {
		return goModuleInfo{}, errNoSharpPrefix
	}
	if substrings[0] == "##" {
		return goModuleInfo{}, errComment
	}
	switch len(substrings) {
	case 3:
		modInfo := goModuleInfo{
			name:          substrings[1],
			pseudoVersion: substrings[2],
		}
		return modInfo, nil
	case 5:
		if substrings[2] != "=>" {
			return goModuleInfo{}, fmt.Errorf("Parse error - unknown format")
		}
		modInfo := goModuleInfo{
			name:          substrings[3],
			pseudoVersion: substrings[4],
		}
		return modInfo, nil
	case 6:
		if substrings[3] != "=>" {
			return goModuleInfo{}, fmt.Errorf("Parse error - unknown format")
		}
		modInfo := goModuleInfo{
			name:          substrings[4],
			pseudoVersion: substrings[5],
		}
		return modInfo, nil
	default:
		return goModuleInfo{}, fmt.Errorf("Parse error - incorrect number of substrings (%d - expected %d)", len(substrings), 3)
	}
}

func pseudoVersionToRpmVersion(pseudoVersion string) (string, error) {
	// https://golang.org/ref/mod#pseudo-versions

	versionRegexp := regexp.MustCompile("^v[0-9]+.[0-9]+.[0-9]+$")
	versionIncompatibleRegexp := regexp.MustCompile(`^v[0-9]+.[0-9]+.[0-9]+\+incompatible$`)
	dateRegexp := regexp.MustCompile("^([a-z]*.)?(?:0.)*([0-9]{8})[0-9]{6}$")
	commitRegexp := regexp.MustCompile("^[0-9a-f]{12}$")

	substrings := strings.Split(pseudoVersion, "-")
	switch len(substrings) {
	case 1:
		// v1.2.3
		if versionRegexp.MatchString(substrings[0]) {
			return strings.TrimPrefix(substrings[0], "v"), nil
		}
		// v1.2.3+incompatible
		if versionIncompatibleRegexp.MatchString(substrings[0]) {
			return strings.TrimSuffix(strings.TrimPrefix(substrings[0], "v"), "+incompatible"), nil
		}
		return "", fmt.Errorf("Failed to parse version substring: %s", substrings[0])

	case 3:
		// v1.2.3-202011221345-0123456789abc
		// v1.1.2-0.20210202002709-95e28344e08c
		// v0.0.0-alpha.0.0.20201126035554-299b6af535d1
		if !versionRegexp.MatchString(substrings[0]) {
			return "", fmt.Errorf("Failed to parse version substring: %s", substrings[0])
		}
		version := strings.TrimPrefix(substrings[0], "v")

		dateSubstrings := dateRegexp.FindStringSubmatch(substrings[1])
		if dateSubstrings == nil || len(dateSubstrings) != 3 {
			return "", fmt.Errorf("Failed to parse date substring: %s", substrings[1])
		}
		extraver := dateSubstrings[1]
		date := dateSubstrings[2]

		if !commitRegexp.MatchString(substrings[2]) {
			return "", fmt.Errorf("Failed to parse commit substring: %s", substrings[2])
		}
		commit := substrings[2]
		snapinfo := fmt.Sprintf("%sgit%s", date, commit)

		return fmt.Sprintf("%s-0.%s%s", version, extraver, snapinfo), nil
	default:
		return "", fmt.Errorf("Failed to parse pseudoversion")
	}
}

func printBundledProvides(vendoredModules []goModuleInfo) {
	sort.Sort(byModuleName(vendoredModules))
	for _, modInfo := range vendoredModules {
		rpmVersion, err := pseudoVersionToRpmVersion(modInfo.pseudoVersion)
		if err == nil {
			fmt.Printf("Provides: bundled(golang(%s)) = %s\n", modInfo.name, rpmVersion)
		} else {
			fmt.Fprintf(os.Stderr, "failed to parse pseudoversion %s for module %s: %v\n", modInfo.pseudoVersion, modInfo.name, err)
			fmt.Printf("Provides: bundled(golang(%s))\n", modInfo.name)
		}
	}
}

func fetchVendoredModules() ([]goModuleInfo, error) {
	scanner := modScanner(govendorScannerNew())

	vendoredModules := []goModuleInfo{}

	for scanner.Scan() {
		line := scanner.Text()
		modInfo, err := parseModuleLine(line)
		if err != nil {
			if errors.Is(err, errNoSharpPrefix) || errors.Is(err, errComment) || errors.Is(err, errEmptyLine) {
				continue
			}
			fmt.Fprintf(os.Stderr, "failed to parse line: %v\n\t%s\n", err, line)
			continue
		}
		vendoredModules = append(vendoredModules, modInfo)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if err := scanner.Close(); err != nil {
		return nil, err
	}
	return vendoredModules, nil
}

func main() {
	vendoredModules, err := fetchVendoredModules()
	if err != nil {
		log.Fatal(err)
	}

	printBundledProvides(vendoredModules)
}
