package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/f1nniboy/chorus/internal/meta"
)

type release struct {
	when    time.Time
	version string
}

type xmlURL struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type xmlRelease struct {
	XMLName xml.Name `xml:"release"`
	Version string   `xml:"version,attr"`
	Date    string   `xml:"date,attr"`
	URL     xmlURL   `xml:"url"`
}

func main() {
	version := flag.String("version", "", "pending release version, without a leading v")
	flag.Parse()
	if *version == "" {
		log.Fatal("metainfogen: -version is required")
	}

	releases, err := gitReleases()
	if err != nil {
		log.Fatalf("metainfogen: %v", err)
	}
	releases = append(releases, release{version: *version, when: time.Now()})

	slices.SortFunc(releases, func(a, b release) int {
		return b.when.Compare(a.when)
	})

	block, err := buildReleasesBlock(releases)
	if err != nil {
		log.Fatalf("metainfogen: %v", err)
	}

	path := "data/" + meta.AppID + ".metainfo.xml"
	if err := spliceReleases(path, block); err != nil {
		log.Fatalf("metainfogen: %v", err)
	}
}

func gitReleases() ([]release, error) {
	out, err := exec.Command("git", "tag").Output()
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	var releases []release
	for tag := range strings.FieldsSeq(string(out)) {
		version, ok := strings.CutPrefix(tag, "v")
		if !ok {
			continue
		}

		dateOut, err := exec.Command("git", "log", "-1", "--format=%aI", tag).Output()
		if err != nil {
			return nil, fmt.Errorf("date for tag %s: %w", tag, err)
		}
		when, err := time.Parse(time.RFC3339, strings.TrimSpace(string(dateOut)))
		if err != nil {
			return nil, fmt.Errorf("date for tag %s: %w", tag, err)
		}

		releases = append(releases, release{version: version, when: when})
	}
	return releases, nil
}

func buildReleasesBlock(releases []release) (string, error) {
	var lines []string
	for _, r := range releases {
		xr := xmlRelease{
			Version: r.version,
			Date:    r.when.Format("2006-01-02"),
			URL:     xmlURL{Type: "details", Value: meta.AppRepo + "/releases/tag/v" + r.version},
		}
		out, err := xml.MarshalIndent(xr, "    ", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal release %s: %w", r.version, err)
		}
		lines = append(lines, string(out))
	}
	return strings.Join(lines, "\n"), nil
}

func spliceReleases(path, block string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	data := string(raw)

	const open, closeTag = "  <releases>\n", "\n  </releases>"
	start := strings.Index(data, open)
	if start == -1 {
		return fmt.Errorf("%s: no %q found", path, strings.TrimSpace(open))
	}
	start += len(open)
	end := strings.Index(data[start:], closeTag)
	if end == -1 {
		return fmt.Errorf("%s: no closing %q found after <releases>", path, strings.TrimSpace(closeTag))
	}
	end += start

	return os.WriteFile(path, []byte(data[:start]+block+data[end:]), 0o644)
}
