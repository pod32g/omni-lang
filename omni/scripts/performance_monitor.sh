#!/bin/bash

# Performance monitoring script for OmniLang
# This script runs performance benchmarks and detects regressions

set -e

# Configuration
PERF_DIR="performance"
BASELINE_FILE="baseline_performance.json"
CURRENT_FILE="current_performance.json"
REGRESSION_THRESHOLD=0.20  # 20% regression threshold

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if baseline exists
check_baseline() {
    if [ ! -f "${PERF_DIR}/${BASELINE_FILE}" ]; then
        print_status $YELLOW "No baseline performance data found. Creating baseline..."
        create_baseline
    fi
}

# Function to create baseline performance data
create_baseline() {
    print_status $GREEN "Creating baseline performance data..."
    
    # Create performance directory if it doesn't exist
    mkdir -p "${PERF_DIR}"
    
    # Run benchmarks and save as baseline
    go test -bench=. -benchmem ./internal/compiler/ -run=^$ > "${PERF_DIR}/benchmark_output.txt" 2>&1
    
    # Extract performance data and save as JSON
    python3 -c "
import json
import re
import os

# Parse benchmark output and create baseline
baseline = {
    'timestamp': '$(date -u +"%Y-%m-%dT%H:%M:%SZ")',
    'platform': '$(uname -s)',
    'architecture': '$(uname -m)',
    'go_version': '$(go version)',
    'tests': {}
}

# Read benchmark output
with open('${PERF_DIR}/benchmark_output.txt', 'r') as f:
    content = f.read()

# Parse benchmark results
pattern = r'Benchmark(\w+)\s+(\d+)\s+(\d+\.?\d*)\s+ns/op\s+(\d+)\s+(\d+\.?\d*)\s+B/op\s+(\d+)\s+allocs/op'
matches = re.findall(pattern, content)

for match in matches:
    name, iterations, ns_per_op, b_per_op, allocs_per_op = match
    baseline['tests'][name] = {
        'name': name,
        'duration_ns': float(ns_per_op) * int(iterations),
        'memory_usage_bytes': int(b_per_op),
        'iterations': int(iterations),
        'throughput': int(iterations) / (float(ns_per_op) * int(iterations) / 1e9)
    }

# Save baseline
with open('${PERF_DIR}/${BASELINE_FILE}', 'w') as f:
    json.dump(baseline, f, indent=2)

print('Baseline created with', len(baseline['tests']), 'tests')
"
    
    print_status $GREEN "Baseline created successfully!"
}

# Function to run performance tests
run_performance_tests() {
    print_status $GREEN "Running performance tests..."
    
    # Run benchmarks
    go test -bench=. -benchmem ./internal/compiler/ -run=^$ > "${PERF_DIR}/current_benchmark_output.txt" 2>&1
    
    # Extract current performance data
    python3 -c "
import json
import re
import os

# Parse benchmark output and create current metrics
current = {
    'timestamp': '$(date -u +"%Y-%m-%dT%H:%M:%SZ")',
    'platform': '$(uname -s)',
    'architecture': '$(uname -m)',
    'go_version': '$(go version)',
    'tests': {}
}

# Read benchmark output
with open('${PERF_DIR}/current_benchmark_output.txt', 'r') as f:
    content = f.read()

# Parse benchmark results
pattern = r'Benchmark(\w+)\s+(\d+)\s+(\d+\.?\d*)\s+ns/op\s+(\d+)\s+(\d+\.?\d*)\s+B/op\s+(\d+)\s+allocs/op'
matches = re.findall(pattern, content)

for match in matches:
    name, iterations, ns_per_op, b_per_op, allocs_per_op = match
    current['tests'][name] = {
        'name': name,
        'duration_ns': float(ns_per_op) * int(iterations),
        'memory_usage_bytes': int(b_per_op),
        'iterations': int(iterations),
        'throughput': int(iterations) / (float(ns_per_op) * int(iterations) / 1e9)
    }

# Save current metrics
with open('${PERF_DIR}/${CURRENT_FILE}', 'w') as f:
    json.dump(current, f, indent=2)

print('Current metrics saved with', len(current['tests']), 'tests')
"
    
    print_status $GREEN "Performance tests completed!"
}

# Function to detect performance regressions
detect_regressions() {
    print_status $GREEN "Detecting performance regressions..."
    
    python3 -c "
import json
import sys

# Load baseline and current metrics
with open('${PERF_DIR}/${BASELINE_FILE}', 'r') as f:
    baseline = json.load(f)

with open('${PERF_DIR}/${CURRENT_FILE}', 'r') as f:
    current = json.load(f)

regressions = []
improvements = []

# Compare metrics
for test_name, current_test in current['tests'].items():
    if test_name in baseline['tests']:
        baseline_test = baseline['tests'][test_name]
        
        # Calculate performance change
        duration_change = (current_test['duration_ns'] - baseline_test['duration_ns']) / baseline_test['duration_ns']
        memory_change = (current_test['memory_usage_bytes'] - baseline_test['memory_usage_bytes']) / baseline_test['memory_usage_bytes']
        
        # Check for regressions (threshold: ${REGRESSION_THRESHOLD})
        if duration_change > ${REGRESSION_THRESHOLD}:
            regressions.append({
                'test': test_name,
                'duration_change': duration_change,
                'memory_change': memory_change,
                'baseline_duration': baseline_test['duration_ns'],
                'current_duration': current_test['duration_ns']
            })
        elif duration_change < -0.05:  # 5% improvement
            improvements.append({
                'test': test_name,
                'duration_change': duration_change,
                'memory_change': memory_change,
                'baseline_duration': baseline_test['duration_ns'],
                'current_duration': current_test['duration_ns']
            })

# Print results
if regressions:
    print('\\n🚨 PERFORMANCE REGRESSIONS DETECTED:')
    for reg in regressions:
        print(f'  {reg[\"test\"]}: {reg[\"duration_change\"]*100:.1f}% slower')
        print(f'    Baseline: {reg[\"baseline_duration\"]:.0f}ns, Current: {reg[\"current_duration\"]:.0f}ns')
    sys.exit(1)
else:
    print('[OK] No performance regressions detected!')

if improvements:
    print('\\n[IMPROVEMENTS] PERFORMANCE IMPROVEMENTS:')
    for imp in improvements:
        print(f'  {imp[\"test\"]}: {imp[\"duration_change\"]*100:.1f}% faster')
"
}

# Function to generate performance report
generate_report() {
    print_status $GREEN "Generating performance report..."
    
    python3 -c "
import json
import os
from datetime import datetime

# Load metrics
with open('${PERF_DIR}/${BASELINE_FILE}', 'r') as f:
    baseline = json.load(f)

with open('${PERF_DIR}/${CURRENT_FILE}', 'r') as f:
    current = json.load(f)

# Generate HTML report
html = f'''
<!DOCTYPE html>
<html>
<head>
    <title>OmniLang Performance Report</title>
    <style>
        body {{ font-family: Arial, sans-serif; margin: 20px; }}
        .header {{ background-color: #f0f0f0; padding: 20px; border-radius: 5px; }}
        .test {{ margin: 10px 0; padding: 10px; border: 1px solid #ddd; border-radius: 5px; }}
        .regression {{ background-color: #ffebee; border-color: #f44336; }}
        .improvement {{ background-color: #e8f5e8; border-color: #4caf50; }}
        .neutral {{ background-color: #f9f9f9; }}
        table {{ border-collapse: collapse; width: 100%; }}
        th, td {{ border: 1px solid #ddd; padding: 8px; text-align: left; }}
        th {{ background-color: #f2f2f2; }}
    </style>
</head>
<body>
    <div class=\"header\">
        <h1>OmniLang Performance Report</h1>
        <p>Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>
        <p>Platform: {current['platform']} {current['architecture']}</p>
        <p>Go Version: {current['go_version']}</p>
    </div>
    
    <h2>Performance Summary</h2>
    <table>
        <tr>
            <th>Test</th>
            <th>Baseline (ns)</th>
            <th>Current (ns)</th>
            <th>Change</th>
            <th>Memory (bytes)</th>
        </tr>
'''

# Add test results
for test_name, current_test in current['tests'].items():
    if test_name in baseline['tests']:
        baseline_test = baseline['tests'][test_name]
        duration_change = (current_test['duration_ns'] - baseline_test['duration_ns']) / baseline_test['duration_ns']
        
        change_class = 'neutral'
        if duration_change > ${REGRESSION_THRESHOLD}:
            change_class = 'regression'
        elif duration_change < -0.05:
            change_class = 'improvement'
        
        html += f'''
        <tr class=\"{change_class}\">
            <td>{test_name}</td>
            <td>{baseline_test['duration_ns']:.0f}</td>
            <td>{current_test['duration_ns']:.0f}</td>
            <td>{duration_change*100:.1f}%</td>
            <td>{current_test['memory_usage_bytes']}</td>
        </tr>
        '''

html += '''
    </table>
</body>
</html>
'''

# Save report
with open('${PERF_DIR}/performance_report.html', 'w') as f:
    f.write(html)

print('Performance report generated: ${PERF_DIR}/performance_report.html')
"
    
    print_status $GREEN "Performance report generated: ${PERF_DIR}/performance_report.html"
}

# Function to clean up old performance data
cleanup() {
    print_status $YELLOW "Cleaning up old performance data..."
    
    # Keep only the last 10 performance runs
    ls -t "${PERF_DIR}"/current_performance_*.json 2>/dev/null | tail -n +11 | xargs rm -f 2>/dev/null || true
    
    print_status $GREEN "Cleanup completed!"
}

# Main execution
main() {
    print_status $GREEN "Starting OmniLang performance monitoring..."
    
    # Check if we're in the right directory
    if [ ! -f "go.mod" ]; then
        print_status $RED "Error: Please run this script from the OmniLang project root directory"
        exit 1
    fi
    
    # Check if Go is available
    if ! command -v go &> /dev/null; then
        print_status $RED "Error: Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if Python is available
    if ! command -v python3 &> /dev/null; then
        print_status $RED "Error: Python 3 is not installed or not in PATH"
        exit 1
    fi
    
    # Parse command line arguments
    case "${1:-run}" in
        "baseline")
            create_baseline
            ;;
        "run")
            check_baseline
            run_performance_tests
            detect_regressions
            generate_report
            cleanup
            ;;
        "report")
            generate_report
            ;;
        "clean")
            cleanup
            ;;
        *)
            echo "Usage: $0 [baseline|run|report|clean]"
            echo "  baseline: Create baseline performance data"
            echo "  run:      Run performance tests and detect regressions (default)"
            echo "  report:   Generate performance report"
            echo "  clean:    Clean up old performance data"
            exit 1
            ;;
    esac
    
    print_status $GREEN "Performance monitoring completed!"
}

# Run main function
main "$@"
