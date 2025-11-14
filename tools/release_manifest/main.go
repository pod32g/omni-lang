package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type artifact struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	SHA  string `json:"sha256"`
}

type manifest struct {
	Version   string     `json:"version"`
	Generated string     `json:"generated"`
	Artifacts []artifact `json:"artifacts"`
}

func main() {
	dir := flag.String("dir", "", "directory containing release artifacts")
	version := flag.String("version", "dev", "release version string")
	output := flag.String("output", "release.json", "manifest output file")
	flag.Parse()

	if *dir == "" {
		fmt.Fprintln(os.Stderr, "--dir is required")
		os.Exit(2)
	}

	entries, err := collectArtifacts(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect artifacts: %v\n", err)
		os.Exit(1)
	}

	data := manifest{
		Version:   *version,
		Generated: time.Now().UTC().Format(time.RFC3339),
		Artifacts: entries,
	}

	file, err := os.Create(filepath.Join(*dir, *output))
	if err != nil {
		fmt.Fprintf(os.Stderr, "create manifest: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "encode manifest: %v\n", err)
		os.Exit(1)
	}
}

func collectArtifacts(dir string) ([]artifact, error) {
	var artifacts []artifact
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := filepath.Base(path)
		if name == "release.json" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		sum, err := sha256File(path)
		if err != nil {
			return err
		}
		artifacts = append(artifacts, artifact{
			Name: name,
			Size: info.Size(),
			SHA:  sum,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return artifacts, nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
