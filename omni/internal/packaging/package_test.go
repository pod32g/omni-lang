package packaging

import (
	"testing"
)

func TestCreatePackage(t *testing.T) {
	// Test creating a package with config - expect error due to missing files
	config := PackageConfig{
		OutputPath:   "test.tar.gz",
		PackageType:  PackageTypeTarGz,
		IncludeDebug: false,
		IncludeSrc:   false,
		Version:      "1.0.0",
		Platform:     "linux",
		Architecture: "amd64",
	}

	err := CreatePackage(config)
	// We expect an error because the required files don't exist
	if err == nil {
		t.Error("Expected CreatePackage to fail with missing files")
	}
}

func TestCreateTarGzPackage(t *testing.T) {
	// Test creating a tar.gz package - expect error due to missing files
	config := PackageConfig{
		OutputPath:   "test.tar.gz",
		PackageType:  PackageTypeTarGz,
		IncludeDebug: false,
		IncludeSrc:   false,
		Version:      "1.0.0",
		Platform:     "linux",
		Architecture: "amd64",
	}

	err := createTarGzPackage(config)
	// We expect an error because the required files don't exist
	if err == nil {
		t.Error("Expected createTarGzPackage to fail with missing files")
	}
}

func TestCreateZipPackage(t *testing.T) {
	// Test creating a zip package - expect error due to missing files
	config := PackageConfig{
		OutputPath:   "test.zip",
		PackageType:  PackageTypeZip,
		IncludeDebug: false,
		IncludeSrc:   false,
		Version:      "1.0.0",
		Platform:     "linux",
		Architecture: "amd64",
	}

	err := createZipPackage(config)
	// We expect an error because the required files don't exist
	if err == nil {
		t.Error("Expected createZipPackage to fail with missing files")
	}
}

func TestGetDefaultPackageName(t *testing.T) {
	// Test getting default package name
	name := GetDefaultPackageName("1.0.0", "linux", "amd64", PackageTypeTarGz)
	if name == "" {
		t.Error("Expected non-empty default package name")
	}

	expected := "omni-lang-1.0.0-linux-amd64.tar.gz"
	if name != expected {
		t.Errorf("Expected package name '%s', got '%s'", expected, name)
	}
}

func TestPackageTypeConstants(t *testing.T) {
	// Test package type constants
	if PackageTypeTarGz != "tar.gz" {
		t.Errorf("Expected PackageTypeTarGz to be 'tar.gz', got '%s'", PackageTypeTarGz)
	}

	if PackageTypeZip != "zip" {
		t.Errorf("Expected PackageTypeZip to be 'zip', got '%s'", PackageTypeZip)
	}
}

func TestPackageConfig(t *testing.T) {
	// Test package config creation
	config := PackageConfig{
		OutputPath:   "test.tar.gz",
		PackageType:  PackageTypeTarGz,
		IncludeDebug: true,
		IncludeSrc:   true,
		Version:      "1.0.0",
		Platform:     "linux",
		Architecture: "amd64",
	}

	if config.OutputPath != "test.tar.gz" {
		t.Errorf("Expected OutputPath 'test.tar.gz', got '%s'", config.OutputPath)
	}

	if config.PackageType != PackageTypeTarGz {
		t.Errorf("Expected PackageType PackageTypeTarGz, got '%s'", config.PackageType)
	}

	if !config.IncludeDebug {
		t.Error("Expected IncludeDebug to be true")
	}

	if !config.IncludeSrc {
		t.Error("Expected IncludeSrc to be true")
	}

	if config.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", config.Version)
	}

	if config.Platform != "linux" {
		t.Errorf("Expected Platform 'linux', got '%s'", config.Platform)
	}

	if config.Architecture != "amd64" {
		t.Errorf("Expected Architecture 'amd64', got '%s'", config.Architecture)
	}
}
