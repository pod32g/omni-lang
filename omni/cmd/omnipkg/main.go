package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/omni-lang/omni/internal/packaging"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	var (
		output       = flag.String("o", "", "output package path")
		packageType  = flag.String("type", "tar.gz", "package type (tar.gz|zip)")
		includeDebug = flag.Bool("debug", false, "include debug symbols")
		includeSrc   = flag.Bool("src", false, "include source code")
		version      = flag.String("version", "dev", "package version")
		platform     = flag.String("platform", runtime.GOOS, "target platform")
		arch         = flag.String("arch", runtime.GOARCH, "target architecture")
		help         = flag.Bool("help", false, "show help and exit")
		showHelp     = flag.Bool("h", false, "show help and exit")
	)

	flag.Parse()

	if *help || *showHelp {
		showUsage()
		return
	}

	// Determine package type
	var pkgType packaging.PackageType
	switch *packageType {
	case "tar.gz":
		pkgType = packaging.PackageTypeTarGz
	case "zip":
		pkgType = packaging.PackageTypeZip
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported package type: %s\n", *packageType)
		fmt.Fprintf(os.Stderr, "Supported types: tar.gz, zip\n")
		os.Exit(1)
	}

	// Generate output path if not provided
	outputPath := *output
	if outputPath == "" {
		outputPath = packaging.GetDefaultPackageName(*version, *platform, *arch, pkgType)
	}

	// Create package configuration
	config := packaging.PackageConfig{
		OutputPath:   outputPath,
		PackageType:  pkgType,
		IncludeDebug: *includeDebug,
		IncludeSrc:   *includeSrc,
		Version:      *version,
		Platform:     *platform,
		Architecture: *arch,
	}

	fmt.Printf("Creating package: %s\n", outputPath)
	fmt.Printf("  Type: %s\n", *packageType)
	fmt.Printf("  Version: %s\n", *version)
	fmt.Printf("  Platform: %s-%s\n", *platform, *arch)
	fmt.Printf("  Include Debug: %t\n", *includeDebug)
	fmt.Printf("  Include Source: %t\n", *includeSrc)
	fmt.Println()

	// Create the package
	if err := packaging.CreatePackage(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating package: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Package created successfully: %s\n", outputPath)
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "Usage: omnipkg [options]\n\n")
	fmt.Fprintf(os.Stderr, "Create distribution packages for OmniLang\n\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -o string\n")
	fmt.Fprintf(os.Stderr, "        output package path (default: auto-generated)\n")
	fmt.Fprintf(os.Stderr, "  -type string\n")
	fmt.Fprintf(os.Stderr, "        package type (tar.gz|zip) (default \"tar.gz\")\n")
	fmt.Fprintf(os.Stderr, "  -debug\n")
	fmt.Fprintf(os.Stderr, "        include debug symbols\n")
	fmt.Fprintf(os.Stderr, "  -src\n")
	fmt.Fprintf(os.Stderr, "        include source code\n")
	fmt.Fprintf(os.Stderr, "  -version string\n")
	fmt.Fprintf(os.Stderr, "        package version (default \"dev\")\n")
	fmt.Fprintf(os.Stderr, "  -platform string\n")
	fmt.Fprintf(os.Stderr, "        target platform (default current OS)\n")
	fmt.Fprintf(os.Stderr, "  -arch string\n")
	fmt.Fprintf(os.Stderr, "        target architecture (default current arch)\n")
	fmt.Fprintf(os.Stderr, "  -help, -h\n")
	fmt.Fprintf(os.Stderr, "        show help and exit\n\n")
	fmt.Fprintf(os.Stderr, "EXAMPLES:\n")
	fmt.Fprintf(os.Stderr, "  omnipkg                                    # Create tar.gz package\n")
	fmt.Fprintf(os.Stderr, "  omnipkg -type zip                         # Create zip package\n")
	fmt.Fprintf(os.Stderr, "  omnipkg -debug -src                       # Include debug and source\n")
	fmt.Fprintf(os.Stderr, "  omnipkg -o my-package.tar.gz              # Custom output name\n")
	fmt.Fprintf(os.Stderr, "  omnipkg -version 1.0.0 -platform linux    # Specific version and platform\n")
}
