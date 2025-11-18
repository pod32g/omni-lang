package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/omni-lang/omni/tools/coverage"
)

func main() {
	var (
		analyzeCmd   = flag.NewFlagSet("analyze", flag.ExitOnError)
		reportCmd    = flag.NewFlagSet("report", flag.ExitOnError)
		checkCmd     = flag.NewFlagSet("check", flag.ExitOnError)
		formatFlag   = reportCmd.String("format", "text", "output format (text|html)")
		thresholdFlag = checkCmd.Float64("threshold", 60.0, "coverage threshold percentage")
	)

	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "analyze":
		analyzeCmd.Parse(os.Args[2:])
		if analyzeCmd.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "Usage: omnicover analyze <coverage.json> <std-library-path>\n")
			os.Exit(1)
		}
		coveragePath := analyzeCmd.Arg(0)
		stdPath := analyzeCmd.Arg(1)
		if err := runAnalyze(coveragePath, stdPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "report":
		reportCmd.Parse(os.Args[2:])
		if reportCmd.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "Usage: omnicover report <coverage.json> <std-library-path> [--format=text|html]\n")
			os.Exit(1)
		}
		coveragePath := reportCmd.Arg(0)
		stdPath := reportCmd.Arg(1)
		outputPath := ""
		if reportCmd.NArg() > 2 {
			outputPath = reportCmd.Arg(2)
		}
		if err := runReport(coveragePath, stdPath, *formatFlag, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "check":
		checkCmd.Parse(os.Args[2:])
		if checkCmd.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "Usage: omnicover check <coverage.json> <std-library-path> [--threshold=60]\n")
			os.Exit(1)
		}
		coveragePath := checkCmd.Arg(0)
		stdPath := checkCmd.Arg(1)
		if err := runCheck(coveragePath, stdPath, *thresholdFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "OmniLang Coverage Tool\n\n")
	fmt.Fprintf(os.Stderr, "USAGE:\n")
	fmt.Fprintf(os.Stderr, "  omnicover <command> [options] <args>\n\n")
	fmt.Fprintf(os.Stderr, "COMMANDS:\n")
	fmt.Fprintf(os.Stderr, "  analyze <coverage.json> <std-library-path>\n")
	fmt.Fprintf(os.Stderr, "        Analyze coverage data and print summary\n")
	fmt.Fprintf(os.Stderr, "  report <coverage.json> <std-library-path> [--format=text|html] [output]\n")
	fmt.Fprintf(os.Stderr, "        Generate coverage report\n")
	fmt.Fprintf(os.Stderr, "  check <coverage.json> <std-library-path> [--threshold=60]\n")
	fmt.Fprintf(os.Stderr, "        Check if coverage meets threshold\n\n")
}

func runAnalyze(coveragePath, stdPath string) error {
	// Parse coverage data
	coverageData, err := coverage.ParseCoverageFile(coveragePath)
	if err != nil {
		return fmt.Errorf("parse coverage file: %w", err)
	}

	// Parse std library
	funcsByFile, err := coverage.ParseStdLibrary(stdPath)
	if err != nil {
		return fmt.Errorf("parse std library: %w", err)
	}

	// Match coverage to functions
	matches := coverage.MatchCoverageToFunctions(coverageData, funcsByFile)

	// Calculate statistics
	stats := coverage.CalculateCoverage(matches)

	// Print summary
	fmt.Printf("Coverage Analysis\n")
	fmt.Printf("================\n\n")
	fmt.Printf("Total Runtime-Wired Functions: %d\n", stats.TotalFunctions)
	fmt.Printf("Covered Functions: %d\n", stats.CoveredFunctions)
	fmt.Printf("Function Coverage: %.2f%%\n", stats.GetFunctionCoveragePercentage())
	fmt.Printf("Line Coverage: %.2f%%\n", stats.GetLineCoveragePercentage())

	return nil
}

func runReport(coveragePath, stdPath, format, outputPath string) error {
	// Parse coverage data
	coverageData, err := coverage.ParseCoverageFile(coveragePath)
	if err != nil {
		return fmt.Errorf("parse coverage file: %w", err)
	}

	// Parse std library
	funcsByFile, err := coverage.ParseStdLibrary(stdPath)
	if err != nil {
		return fmt.Errorf("parse std library: %w", err)
	}

	// Match coverage to functions
	matches := coverage.MatchCoverageToFunctions(coverageData, funcsByFile)

	// Calculate statistics
	stats := coverage.CalculateCoverage(matches)

	// Generate report
	switch format {
	case "text":
		return coverage.GenerateTextReport(stats, outputPath)
	case "html":
		if outputPath == "" {
			outputPath = "coverage.html"
		}
		return coverage.GenerateHTMLReport(stats, outputPath)
	default:
		return fmt.Errorf("unsupported format: %s (use 'text' or 'html')", format)
	}
}

func runCheck(coveragePath, stdPath string, threshold float64) error {
	// Parse coverage data
	coverageData, err := coverage.ParseCoverageFile(coveragePath)
	if err != nil {
		return fmt.Errorf("parse coverage file: %w", err)
	}

	// Parse std library
	funcsByFile, err := coverage.ParseStdLibrary(stdPath)
	if err != nil {
		return fmt.Errorf("parse std library: %w", err)
	}

	// Match coverage to functions
	matches := coverage.MatchCoverageToFunctions(coverageData, funcsByFile)

	// Calculate statistics
	stats := coverage.CalculateCoverage(matches)

	// Check threshold
	meetsThreshold, message := coverage.CheckCoverageThreshold(stats, threshold)
	fmt.Println(message)

	if !meetsThreshold {
		os.Exit(1)
	}

	return nil
}

