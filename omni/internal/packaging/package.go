package packaging

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// PackageType represents the type of distribution package
type PackageType string

const (
	PackageTypeTarGz PackageType = "tar.gz"
	PackageTypeZip   PackageType = "zip"
)

// PackageEntry captures a single artifact within a package.
type PackageEntry struct {
	Source  string `json:"source"`
	Archive string `json:"archive"`
	Size    int64  `json:"size"`
}

type recorderFunc func(source, archive string, size int64)

// PackageConfig holds configuration for creating distribution packages
type PackageConfig struct {
	OutputPath   string
	PackageType  PackageType
	IncludeDebug bool
	IncludeSrc   bool
	Version      string
	Platform     string
	Architecture string
	DryRun       bool
	ManifestPath string
	Checksum     bool
	ChecksumPath string
}

// CreatePackage creates a distribution package with the runtime and compiler
func CreatePackage(config PackageConfig) error {
	entries := make([]PackageEntry, 0, 128)
	recorder := func(src, archive string, size int64) {
		entries = append(entries, PackageEntry{
			Source:  src,
			Archive: archive,
			Size:    size,
		})
	}

	if config.DryRun {
		if err := simulatePackage(config, recorder); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "[dry-run] %s would include %d entries\n", config.PackageType, len(entries))
		if config.ManifestPath != "" {
			if err := writeManifest(config.ManifestPath, entries); err != nil {
				return err
			}
		}
		if config.Checksum {
			return fmt.Errorf("checksum requested but package is built in dry-run mode")
		}
		return nil
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(config.OutputPath), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	var err error
	switch config.PackageType {
	case PackageTypeTarGz:
		err = createTarGzPackage(config, recorder)
	case PackageTypeZip:
		err = createZipPackage(config, recorder)
	default:
		err = fmt.Errorf("unsupported package type: %s", config.PackageType)
	}
	if err != nil {
		return err
	}

	if config.ManifestPath != "" {
		if err := writeManifest(config.ManifestPath, entries); err != nil {
			return err
		}
	}

	if config.Checksum {
		path := config.ChecksumPath
		if path == "" {
			path = config.OutputPath + ".sha256"
		}
		if err := writeChecksum(path, config.OutputPath); err != nil {
			return err
		}
	}

	return nil
}

func simulatePackage(config PackageConfig, record recorderFunc) error {
	switch config.PackageType {
	case PackageTypeTarGz:
		gzWriter := gzip.NewWriter(io.Discard)
		defer gzWriter.Close()
		tarWriter := tar.NewWriter(gzWriter)
		defer tarWriter.Close()
		return addFilesToTar(tarWriter, config, record)
	case PackageTypeZip:
		zipWriter := zip.NewWriter(io.Discard)
		defer zipWriter.Close()
		return addFilesToZip(zipWriter, config, record)
	default:
		return fmt.Errorf("unsupported package type: %s", config.PackageType)
	}
}

func writeManifest(path string, entries []PackageEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create manifest directory: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create manifest: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(entries); err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	return nil
}

func writeChecksum(path, packagePath string) error {
	file, err := os.Open(packagePath)
	if err != nil {
		return fmt.Errorf("open package for checksum: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("compute checksum: %w", err)
	}

	sum := hasher.Sum(nil)
	if err := os.WriteFile(path, []byte(fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum), filepath.Base(packagePath))), 0o644); err != nil {
		return fmt.Errorf("write checksum: %w", err)
	}
	return nil
}

// createTarGzPackage creates a tar.gz distribution package
func createTarGzPackage(config PackageConfig, record recorderFunc) error {
	file, err := os.Create(config.OutputPath)
	if err != nil {
		return fmt.Errorf("create package file: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add files to the package
	if err := addFilesToTar(tarWriter, config, record); err != nil {
		return fmt.Errorf("add files to tar: %w", err)
	}

	return nil
}

// createZipPackage creates a zip distribution package
func createZipPackage(config PackageConfig, record recorderFunc) error {
	file, err := os.Create(config.OutputPath)
	if err != nil {
		return fmt.Errorf("create package file: %w", err)
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Add files to the package
	if err := addFilesToZip(zipWriter, config, record); err != nil {
		return fmt.Errorf("add files to zip: %w", err)
	}

	return nil
}

// addFilesToTar adds files to a tar archive
func addFilesToTar(tarWriter *tar.Writer, config PackageConfig, record recorderFunc) error {
	// Add runtime files
	runtimeFiles := []string{
		"runtime/omni_rt.c",
		"runtime/omni_rt.h",
	}

	for _, file := range runtimeFiles {
		if err := addFileToTar(tarWriter, file, "omni-lang-"+config.Version+"/"+file, record); err != nil {
			return err
		}
	}

	// Add compiler binary
	compilerPath := "bin/omnic"
	if err := addFileToTar(tarWriter, compilerPath, "omni-lang-"+config.Version+"/bin/omnic", record); err != nil {
		return err
	}

	// Add runner binary if it exists
	runnerPath := "bin/omnir"
	if _, err := os.Stat(runnerPath); err == nil {
		if err := addFileToTar(tarWriter, runnerPath, "omni-lang-"+config.Version+"/bin/omnir", record); err != nil {
			return err
		}
	}

	// Add standard library
	stdLibDir := "std"
	if err := addDirectoryToTar(tarWriter, stdLibDir, "omni-lang-"+config.Version+"/std", record); err != nil {
		return err
	}

	// Add examples
	examplesDir := "examples"
	if err := addDirectoryToTar(tarWriter, examplesDir, "omni-lang-"+config.Version+"/examples", record); err != nil {
		return err
	}

	// Add documentation
	docFiles := []string{
		"README.md",
		"LICENSE",
		"docs/README.md",
		"docs/quick-reference.md",
	}

	for _, docFile := range docFiles {
		if _, err := os.Stat(docFile); err == nil {
			if err := addFileToTar(tarWriter, docFile, "omni-lang-"+config.Version+"/"+docFile, record); err != nil {
				return err
			}
		}
	}

	// Add debug symbols if requested
	if config.IncludeDebug {
		debugDir := "debug"
		if err := addDirectoryToTar(tarWriter, debugDir, "omni-lang-"+config.Version+"/debug", record); err != nil {
			// Debug directory might not exist, that's okay
		}
	}

	// Add source code if requested
	if config.IncludeSrc {
		srcFiles := []string{
			"internal/",
			"cmd/",
			"go.mod",
			"Makefile",
		}

		for _, srcFile := range srcFiles {
			if err := addFileOrDirToTar(tarWriter, srcFile, "omni-lang-"+config.Version+"/src/"+srcFile, record); err != nil {
				return err
			}
		}
	}

	return nil
}

// addFilesToZip adds files to a zip archive
func addFilesToZip(zipWriter *zip.Writer, config PackageConfig, record recorderFunc) error {
	// Add runtime files
	runtimeFiles := []string{
		"runtime/omni_rt.c",
		"runtime/omni_rt.h",
	}

	for _, file := range runtimeFiles {
		if err := addFileToZip(zipWriter, file, "omni-lang-"+config.Version+"/"+file, record); err != nil {
			return err
		}
	}

	// Add compiler binary
	compilerPath := "bin/omnic"
	if err := addFileToZip(zipWriter, compilerPath, "omni-lang-"+config.Version+"/bin/omnic", record); err != nil {
		return err
	}

	// Add runner binary if it exists
	runnerPath := "bin/omnir"
	if _, err := os.Stat(runnerPath); err == nil {
		if err := addFileToZip(zipWriter, runnerPath, "omni-lang-"+config.Version+"/bin/omnir", record); err != nil {
			return err
		}
	}

	// Add standard library
	stdLibDir := "std"
	if err := addDirectoryToZip(zipWriter, stdLibDir, "omni-lang-"+config.Version+"/std", record); err != nil {
		return err
	}

	// Add examples
	examplesDir := "examples"
	if err := addDirectoryToZip(zipWriter, examplesDir, "omni-lang-"+config.Version+"/examples", record); err != nil {
		return err
	}

	// Add documentation
	docFiles := []string{
		"README.md",
		"LICENSE",
		"docs/README.md",
		"docs/quick-reference.md",
	}

	for _, docFile := range docFiles {
		if _, err := os.Stat(docFile); err == nil {
			if err := addFileToZip(zipWriter, docFile, "omni-lang-"+config.Version+"/"+docFile, record); err != nil {
				return err
			}
		}
	}

	return nil
}

// addFileToTar adds a single file to a tar archive
func addFileToTar(tarWriter *tar.Writer, filePath, archivePath string, record recorderFunc) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name: archivePath,
		Size: stat.Size(),
		Mode: int64(stat.Mode()),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tarWriter, file)
	if err == nil && record != nil {
		record(filePath, archivePath, stat.Size())
	}
	return err
}

// addFileToZip adds a single file to a zip archive
func addFileToZip(zipWriter *zip.Writer, filePath, archivePath string, record recorderFunc) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(stat)
	if err != nil {
		return err
	}

	header.Name = archivePath
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	if err == nil && record != nil {
		record(filePath, archivePath, stat.Size())
	}
	return err
}

// addDirectoryToTar adds a directory and its contents to a tar archive
func addDirectoryToTar(tarWriter *tar.Writer, dirPath, archivePath string, record recorderFunc) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		archiveFilePath := filepath.Join(archivePath, relPath)
		archiveFilePath = filepath.ToSlash(archiveFilePath)

		if info.IsDir() {
			header := &tar.Header{
				Name:     archiveFilePath + "/",
				Typeflag: tar.TypeDir,
				Mode:     int64(info.Mode()),
			}
			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}
			if record != nil {
				record(path, archiveFilePath+"/", 0)
			}
			return nil
		}

		return addFileToTar(tarWriter, path, archiveFilePath, record)
	})
}

// addDirectoryToZip adds a directory and its contents to a zip archive
func addDirectoryToZip(zipWriter *zip.Writer, dirPath, archivePath string, record recorderFunc) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		archiveFilePath := filepath.Join(archivePath, relPath)
		archiveFilePath = filepath.ToSlash(archiveFilePath)

		if info.IsDir() {
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = archiveFilePath + "/"
			header.Method = zip.Deflate
			_, err = zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}
			if record != nil {
				record(path, archiveFilePath+"/", 0)
			}
			return err
		}

		return addFileToZip(zipWriter, path, archiveFilePath, record)
	})
}

// addFileOrDirToTar adds a file or directory to a tar archive
func addFileOrDirToTar(tarWriter *tar.Writer, path, archivePath string, record recorderFunc) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return addDirectoryToTar(tarWriter, path, archivePath, record)
	}
	return addFileToTar(tarWriter, path, archivePath, record)
}

// GetDefaultPackageName generates a default package name based on platform and architecture
func GetDefaultPackageName(version, platform, arch string, packageType PackageType) string {
	return fmt.Sprintf("omni-lang-%s-%s-%s.%s", version, platform, arch, packageType)
}
