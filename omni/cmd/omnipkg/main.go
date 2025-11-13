package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/omni-lang/omni/internal/logging"
	"github.com/omni-lang/omni/internal/packaging"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	var (
		output        = flag.String("o", "", "output package path")
		packageType   = flag.String("type", "tar.gz", "package type (tar.gz|zip)")
		packageTypeSh = flag.String("t", "", "alias for -type")
		includeDebug  = flag.Bool("debug", false, "include debug symbols")
		includeDebugS = flag.Bool("g", false, "alias for -debug")
		includeSrc    = flag.Bool("src", false, "include source code")
		includeSrcAlt = flag.Bool("s", false, "alias for -src")
		version       = flag.String("version", "dev", "package version")
		versionShort  = flag.String("v", "", "alias for -version")
		platform      = flag.String("platform", runtime.GOOS, "target platform")
		platformShort = flag.String("p", "", "alias for -platform")
		arch          = flag.String("arch", runtime.GOARCH, "target architecture")
		archShort     = flag.String("a", "", "alias for -arch")
		listTypes     = flag.Bool("list-types", false, "list supported package types and exit")
		listTypesAlt  = flag.Bool("T", false, "alias for -list-types")
		help          = flag.Bool("help", false, "show help and exit")
		showHelp      = flag.Bool("h", false, "show help and exit")
	)

	flag.Parse()

	logger := logging.Logger()
	logging.SetLevel(logging.LevelInfo)

	if *listTypes || *listTypesAlt {
		fmt.Println("Supported package types:")
		fmt.Println("  tar.gz")
		fmt.Println("  zip")
		return
	}

	if *help || *showHelp {
		showUsage()
		return
	}

	if *packageTypeSh != "" {
		*packageType = *packageTypeSh
	}
	if *includeDebugS {
		*includeDebug = true
	}
	if *includeSrcAlt {
		*includeSrc = true
	}
	if *versionShort != "" {
		*version = *versionShort
	}
	if *platformShort != "" {
		*platform = *platformShort
	}
	if *archShort != "" {
		*arch = *archShort
	}

	// Determine package type
	var pkgType packaging.PackageType
	switch *packageType {
	case "tar.gz":
		pkgType = packaging.PackageTypeTarGz
	case "zip":
		pkgType = packaging.PackageTypeZip
	default:
		logger.ErrorFields("unsupported package type",
			logging.String("type", *packageType),
		)
		fmt.Fprintln(os.Stderr, "Supported types: tar.gz, zip")
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

	logger.InfoFields("Creating package",
		logging.String("output", outputPath),
		logging.String("type", *packageType),
		logging.String("version", *version),
		logging.String("platform", *platform),
		logging.String("arch", *arch),
		logging.Bool("include_debug", *includeDebug),
		logging.Bool("include_src", *includeSrc),
	)

	// Create the package
	if err := packaging.CreatePackage(config); err != nil {
		logger.ErrorFields("failed to create package",
			logging.Error("error", err),
			logging.String("output", outputPath),
		)
		os.Exit(1)
	}

	logger.InfoFields("Package created successfully",
		logging.String("output", outputPath),
	)
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "Usage: omnipkg [options]\n\n")
	fmt.Fprintf(os.Stderr, "Create distribution packages for OmniLang\n\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -o string\n")
	fmt.Fprintf(os.Stderr, "        output package path (default: auto-generated)\n")
	fmt.Fprintf(os.Stderr, "  -type, -t string\n")
	fmt.Fprintf(os.Stderr, "        package type (tar.gz|zip) (default \"tar.gz\")\n")
	fmt.Fprintf(os.Stderr, "  -debug, -g\n")
	fmt.Fprintf(os.Stderr, "        include debug symbols\n")
	fmt.Fprintf(os.Stderr, "  -src, -s\n")
	fmt.Fprintf(os.Stderr, "        include source code\n")
	fmt.Fprintf(os.Stderr, "  -version, -v string\n")
	fmt.Fprintf(os.Stderr, "        package version (default \"dev\")\n")
	fmt.Fprintf(os.Stderr, "  -platform, -p string\n")
	fmt.Fprintf(os.Stderr, "        target platform (default current OS)\n")
	fmt.Fprintf(os.Stderr, "  -arch, -a string\n")
	fmt.Fprintf(os.Stderr, "        target architecture (default current arch)\n")
	fmt.Fprintf(os.Stderr, "  -list-types, -T\n")
	fmt.Fprintf(os.Stderr, "        list supported package types and exit\n")
	fmt.Fprintf(os.Stderr, "  -help, -h\n")
	fmt.Fprintf(os.Stderr, "        show help and exit\n\n")
	fmt.Fprintf(os.Stderr, "EXAMPLES:\n")
	fmt.Fprintf(os.Stderr, "  omnipkg                                    # Create tar.gz package\n")
	fmt.Fprintf(os.Stderr, "  omnipkg -type zip                         # Create zip package\n")
	fmt.Fprintf(os.Stderr, "  omnipkg -debug -src                       # Include debug and source\n")
	fmt.Fprintf(os.Stderr, "  omnipkg -o my-package.tar.gz              # Custom output name\n")
	fmt.Fprintf(os.Stderr, "  omnipkg -version 1.0.0 -platform linux    # Specific version and platform\n")
}
