package packaging

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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

// PackageConfig holds configuration for creating distribution packages
type PackageConfig struct {
	OutputPath   string
	PackageType  PackageType
	IncludeDebug bool
	IncludeSrc   bool
	Version      string
	Platform     string
	Architecture string
}

// CreatePackage creates a distribution package with the runtime and compiler
func CreatePackage(config PackageConfig) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(config.OutputPath), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	switch config.PackageType {
	case PackageTypeTarGz:
		return createTarGzPackage(config)
	case PackageTypeZip:
		return createZipPackage(config)
	default:
		return fmt.Errorf("unsupported package type: %s", config.PackageType)
	}
}

// createTarGzPackage creates a tar.gz distribution package
func createTarGzPackage(config PackageConfig) error {
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
	if err := addFilesToTar(tarWriter, config); err != nil {
		return fmt.Errorf("add files to tar: %w", err)
	}

	return nil
}

// createZipPackage creates a zip distribution package
func createZipPackage(config PackageConfig) error {
	file, err := os.Create(config.OutputPath)
	if err != nil {
		return fmt.Errorf("create package file: %w", err)
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Add files to the package
	if err := addFilesToZip(zipWriter, config); err != nil {
		return fmt.Errorf("add files to zip: %w", err)
	}

	return nil
}

// addFilesToTar adds files to a tar archive
func addFilesToTar(tarWriter *tar.Writer, config PackageConfig) error {
	// Add runtime files
	runtimeFiles := []string{
		"runtime/omni_rt.c",
		"runtime/omni_rt.h",
	}

	for _, file := range runtimeFiles {
		if err := addFileToTar(tarWriter, file, "omni-lang-"+config.Version+"/"+file); err != nil {
			return err
		}
	}

	// Add compiler binary
	compilerPath := "bin/omnic"
	if err := addFileToTar(tarWriter, compilerPath, "omni-lang-"+config.Version+"/bin/omnic"); err != nil {
		return err
	}

	// Add runner binary if it exists
	runnerPath := "bin/omnir"
	if _, err := os.Stat(runnerPath); err == nil {
		if err := addFileToTar(tarWriter, runnerPath, "omni-lang-"+config.Version+"/bin/omnir"); err != nil {
			return err
		}
	}

	// Add standard library
	stdLibDir := "std"
	if err := addDirectoryToTar(tarWriter, stdLibDir, "omni-lang-"+config.Version+"/std"); err != nil {
		return err
	}

	// Add examples
	examplesDir := "examples"
	if err := addDirectoryToTar(tarWriter, examplesDir, "omni-lang-"+config.Version+"/examples"); err != nil {
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
			if err := addFileToTar(tarWriter, docFile, "omni-lang-"+config.Version+"/"+docFile); err != nil {
				return err
			}
		}
	}

	// Add debug symbols if requested
	if config.IncludeDebug {
		debugDir := "debug"
		if err := addDirectoryToTar(tarWriter, debugDir, "omni-lang-"+config.Version+"/debug"); err != nil {
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
			if err := addFileOrDirToTar(tarWriter, srcFile, "omni-lang-"+config.Version+"/src/"+srcFile); err != nil {
				return err
			}
		}
	}

	return nil
}

// addFilesToZip adds files to a zip archive
func addFilesToZip(zipWriter *zip.Writer, config PackageConfig) error {
	// Add runtime files
	runtimeFiles := []string{
		"runtime/omni_rt.c",
		"runtime/omni_rt.h",
	}

	for _, file := range runtimeFiles {
		if err := addFileToZip(zipWriter, file, "omni-lang-"+config.Version+"/"+file); err != nil {
			return err
		}
	}

	// Add compiler binary
	compilerPath := "bin/omnic"
	if err := addFileToZip(zipWriter, compilerPath, "omni-lang-"+config.Version+"/bin/omnic"); err != nil {
		return err
	}

	// Add runner binary if it exists
	runnerPath := "bin/omnir"
	if _, err := os.Stat(runnerPath); err == nil {
		if err := addFileToZip(zipWriter, runnerPath, "omni-lang-"+config.Version+"/bin/omnir"); err != nil {
			return err
		}
	}

	// Add standard library
	stdLibDir := "std"
	if err := addDirectoryToZip(zipWriter, stdLibDir, "omni-lang-"+config.Version+"/std"); err != nil {
		return err
	}

	// Add examples
	examplesDir := "examples"
	if err := addDirectoryToZip(zipWriter, examplesDir, "omni-lang-"+config.Version+"/examples"); err != nil {
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
			if err := addFileToZip(zipWriter, docFile, "omni-lang-"+config.Version+"/"+docFile); err != nil {
				return err
			}
		}
	}

	return nil
}

// addFileToTar adds a single file to a tar archive
func addFileToTar(tarWriter *tar.Writer, filePath, archivePath string) error {
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
	return err
}

// addFileToZip adds a single file to a zip archive
func addFileToZip(zipWriter *zip.Writer, filePath, archivePath string) error {
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
	return err
}

// addDirectoryToTar adds a directory and its contents to a tar archive
func addDirectoryToTar(tarWriter *tar.Writer, dirPath, archivePath string) error {
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
			return tarWriter.WriteHeader(header)
		}

		return addFileToTar(tarWriter, path, archiveFilePath)
	})
}

// addDirectoryToZip adds a directory and its contents to a zip archive
func addDirectoryToZip(zipWriter *zip.Writer, dirPath, archivePath string) error {
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
			return err
		}

		return addFileToZip(zipWriter, path, archiveFilePath)
	})
}

// addFileOrDirToTar adds a file or directory to a tar archive
func addFileOrDirToTar(tarWriter *tar.Writer, path, archivePath string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return addDirectoryToTar(tarWriter, path, archivePath)
	}
	return addFileToTar(tarWriter, path, archivePath)
}

// GetDefaultPackageName generates a default package name based on platform and architecture
func GetDefaultPackageName(version, platform, arch string, packageType PackageType) string {
	return fmt.Sprintf("omni-lang-%s-%s-%s.%s", version, platform, arch, packageType)
}
