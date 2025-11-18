package coverage

import (
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"
)

// GenerateTextReport generates a text coverage report
func GenerateTextReport(stats CoverageStats, outputPath string) error {
	var sb strings.Builder

	sb.WriteString("Standard Library Coverage Report\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	funcCoverage := stats.GetFunctionCoveragePercentage()
	lineCoverage := stats.GetLineCoveragePercentage()

	sb.WriteString(fmt.Sprintf("Function Coverage: %.2f%% (%d/%d functions)\n",
		funcCoverage, stats.CoveredFunctions, stats.TotalFunctions))
	sb.WriteString(fmt.Sprintf("Line Coverage: %.2f%% (%d/%d lines)\n\n",
		lineCoverage, stats.CoveredLines, stats.TotalLines))

	// Group by file
	files := make(map[string][]FunctionCoverage)
	for _, fc := range stats.FunctionDetails {
		files[fc.File] = append(files[fc.File], fc)
	}

	// Sort files
	fileList := make([]string, 0, len(files))
	for f := range files {
		fileList = append(fileList, f)
	}
	sort.Strings(fileList)

	for _, file := range fileList {
		funcs := files[file]
		sort.Slice(funcs, func(i, j int) bool {
			return funcs[i].Line < funcs[j].Line
		})

		sb.WriteString(fmt.Sprintf("\n%s\n", file))
		sb.WriteString(strings.Repeat("-", len(file)) + "\n")

		for _, fc := range funcs {
			status := "✗"
			if fc.Covered {
				status = "✓"
			}
			sb.WriteString(fmt.Sprintf("  %s %s (line %d)", status, fc.Function, fc.Line))
			if fc.Covered {
				sb.WriteString(fmt.Sprintf(" - called %d time(s)", fc.CallCount))
			}
			sb.WriteString("\n")
		}
	}

	if outputPath == "" {
		fmt.Print(sb.String())
		return nil
	}

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

// GenerateHTMLReport generates an HTML coverage report
func GenerateHTMLReport(stats CoverageStats, outputPath string) error {
	funcCoverage := stats.GetFunctionCoveragePercentage()
	lineCoverage := stats.GetLineCoveragePercentage()

	// Group by file
	files := make(map[string][]FunctionCoverage)
	for _, fc := range stats.FunctionDetails {
		files[fc.File] = append(files[fc.File], fc)
	}

	// Sort files
	fileList := make([]string, 0, len(files))
	for f := range files {
		fileList = append(fileList, f)
	}
	sort.Strings(fileList)

	// Prepare file data for template
	type FileData struct {
		Path            string
		Functions       []FunctionCoverage
		CoveredCount    int
		TotalCount      int
		CoveragePercent float64
	}

	fileDataList := make([]FileData, 0, len(fileList))
	for _, file := range fileList {
		funcs := files[file]
		sort.Slice(funcs, func(i, j int) bool {
			return funcs[i].Line < funcs[j].Line
		})

		covered := 0
		for _, fc := range funcs {
			if fc.Covered {
				covered++
			}
		}

		coveragePercent := 0.0
		if len(funcs) > 0 {
			coveragePercent = float64(covered) / float64(len(funcs)) * 100.0
		}

		fileDataList = append(fileDataList, FileData{
			Path:            file,
			Functions:       funcs,
			CoveredCount:    covered,
			TotalCount:      len(funcs),
			CoveragePercent: coveragePercent,
		})
	}

	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>OmniLang Standard Library Coverage Report</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            background-color: #f5f5f5;
        }
        .header {
            background-color: #2c3e50;
            color: white;
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        .summary {
            background-color: white;
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .metric {
            display: inline-block;
            margin-right: 30px;
        }
        .metric-value {
            font-size: 24px;
            font-weight: bold;
        }
        .metric-label {
            font-size: 14px;
            color: #666;
        }
        .file {
            background-color: white;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 15px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .file-header {
            font-weight: bold;
            font-size: 16px;
            margin-bottom: 10px;
            color: #2c3e50;
        }
        .function {
            padding: 5px 10px;
            margin: 2px 0;
            border-left: 3px solid #ddd;
        }
        .function.covered {
            border-left-color: #27ae60;
            background-color: #d5f4e6;
        }
        .function.uncovered {
            border-left-color: #e74c3c;
            background-color: #fadbd8;
        }
        .function-name {
            font-weight: bold;
        }
        .function-line {
            color: #666;
            font-size: 12px;
        }
        .coverage-bar {
            height: 20px;
            background-color: #ecf0f1;
            border-radius: 10px;
            overflow: hidden;
            margin: 10px 0;
        }
        .coverage-fill {
            height: 100%;
            background-color: #27ae60;
            transition: width 0.3s;
        }
        .coverage-fill.low {
            background-color: #e74c3c;
        }
        .coverage-fill.medium {
            background-color: #f39c12;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>OmniLang Standard Library Coverage Report</h1>
    </div>
    
    <div class="summary">
        <h2>Summary</h2>
        <div class="metric">
            <div class="metric-value" style="color: {{if ge .FuncCoverage 60.0}}#27ae60{{else if ge .FuncCoverage 30.0}}#f39c12{{else}}#e74c3c{{end}};">
                {{printf "%.2f" .FuncCoverage}}%
            </div>
            <div class="metric-label">Function Coverage</div>
            <div style="font-size: 12px; color: #666;">{{.CoveredFunctions}}/{{.TotalFunctions}} functions</div>
        </div>
        <div class="metric">
            <div class="metric-value" style="color: {{if ge .LineCoverage 60.0}}#27ae60{{else if ge .LineCoverage 30.0}}#f39c12{{else}}#e74c3c{{end}};">
                {{printf "%.2f" .LineCoverage}}%
            </div>
            <div class="metric-label">Line Coverage</div>
            <div style="font-size: 12px; color: #666;">{{.CoveredLines}}/{{.TotalLines}} lines</div>
        </div>
    </div>
    
    <h2>Files</h2>
    {{range .Files}}
    <div class="file">
        <div class="file-header">
            {{.Path}} - {{printf "%.2f" .CoveragePercent}}% ({{.CoveredCount}}/{{.TotalCount}})
        </div>
        <div class="coverage-bar">
            <div class="coverage-fill {{if ge .CoveragePercent 60.0}}{{else if ge .CoveragePercent 30.0}}medium{{else}}low{{end}}" 
                 style="width: {{.CoveragePercent}}%"></div>
        </div>
        {{range .Functions}}
        <div class="function {{if .Covered}}covered{{else}}uncovered{{end}}">
            <span class="function-name">{{.Function}}</span>
            <span class="function-line"> (line {{.Line}})</span>
            {{if .Covered}}
            <span style="color: #27ae60;">✓ Called {{.CallCount}} time(s)</span>
            {{else}}
            <span style="color: #e74c3c;">✗ Not covered</span>
            {{end}}
        </div>
        {{end}}
    </div>
    {{end}}
</body>
</html>`

	t := template.Must(template.New("report").Parse(tmpl))

	data := struct {
		FuncCoverage     float64
		LineCoverage     float64
		CoveredFunctions int
		TotalFunctions   int
		CoveredLines     int
		TotalLines       int
		Files            []FileData
	}{
		FuncCoverage:     funcCoverage,
		LineCoverage:     lineCoverage,
		CoveredFunctions: stats.CoveredFunctions,
		TotalFunctions:   stats.TotalFunctions,
		CoveredLines:     stats.CoveredLines,
		TotalLines:       stats.TotalLines,
		Files:            fileDataList,
	}

	var output strings.Builder
	if err := t.Execute(&output, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	if outputPath == "" {
		fmt.Print(output.String())
		return nil
	}

	return os.WriteFile(outputPath, []byte(output.String()), 0644)
}

// CheckCoverageThreshold checks if coverage meets the threshold
func CheckCoverageThreshold(stats CoverageStats, threshold float64) (bool, string) {
	funcCoverage := stats.GetFunctionCoveragePercentage()
	lineCoverage := stats.GetLineCoveragePercentage()

	meetsThreshold := funcCoverage >= threshold && lineCoverage >= threshold

	message := fmt.Sprintf("Coverage: Function %.2f%%, Line %.2f%% (Threshold: %.2f%%)",
		funcCoverage, lineCoverage, threshold)

	if !meetsThreshold {
		message += " - FAILED"
	} else {
		message += " - PASSED"
	}

	return meetsThreshold, message
}
