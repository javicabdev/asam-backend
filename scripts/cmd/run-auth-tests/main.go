// Package main provides a test runner for authentication system tests
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[0;33m"
	colorRed    = "\033[0;31m"
	colorCyan   = "\033[0;36m"
	colorReset  = "\033[0m"
)

// testConfig holds the configuration for running tests
type testConfig struct {
	testPath        string
	coverageFile    string
	coveragePackage string
	testFiles       []string
}

func main() {
	printHeader()

	// Setup project directory
	if err := setupProjectDirectory(); err != nil {
		printError("Error setting up project directory", err)
		os.Exit(1)
	}

	// Configure test parameters
	config := testConfig{
		testPath:        "./test/internal/domain/services",
		coverageFile:    "coverage_auth.out",
		coveragePackage: "github.com/javicabdev/asam-backend/internal/domain/services",
		testFiles: []string{
			"test/internal/domain/services/auth_service_test.go",
			"test/internal/domain/services/user_service_test.go",
		},
	}

	// Check test files
	if !checkTestFiles(config.testFiles) {
		printError("Some test files are missing!", nil)
		os.Exit(1)
	}

	// Run tests
	duration := runTests(config)

	// Show coverage summary
	showCoverageSummary(config.coverageFile)

	// Generate HTML report
	generateHTMLReport(config.coverageFile)

	// Print final summary
	printFinalSummary(duration)
}

// printHeader prints the test runner header
func printHeader() {
	fmt.Printf("%s🔐 ASAM Backend - Auth System Test Runner%s\n", colorCyan, colorReset)
	fmt.Println(strings.Repeat("=", 50))
}

// setupProjectDirectory sets up the correct working directory
func setupProjectDirectory() error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	// Change to project root if we're in scripts directory
	if strings.HasSuffix(projectRoot, "scripts") {
		projectRoot = filepath.Dir(projectRoot)
		if err := os.Chdir(projectRoot); err != nil {
			return fmt.Errorf("changing directory: %w", err)
		}
	}

	return nil
}

// checkTestFiles verifies that all test files exist
func checkTestFiles(testFiles []string) bool {
	fmt.Printf("\n%s📁 Checking test files...%s\n", colorYellow, colorReset)

	allFilesExist := true
	for _, file := range testFiles {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("  ✅ %s\n", file)
		} else {
			fmt.Printf("  ❌ %s (not found)\n", file)
			allFilesExist = false
		}
	}

	return allFilesExist
}

// runTests executes the test suite
func runTests(config testConfig) time.Duration {
	fmt.Printf("\n%s🧪 Running authentication tests...%s\n", colorYellow, colorReset)

	start := time.Now()

	// Build test command securely by passing each argument separately.

	cmd := exec.Command("go", "test", "-v", "-race",
		"-coverprofile", config.coverageFile,
		"-covermode", "atomic",
		"-coverpkg", config.coveragePackage,
		config.testPath)

	// Set output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run tests
	if err := cmd.Run(); err != nil {
		printError("Tests failed!", nil)
		os.Exit(1)
	}

	duration := time.Since(start)
	fmt.Printf("\n%s✅ All tests passed in %.2fs!%s\n", colorGreen, duration.Seconds(), colorReset)

	return duration
}

// showCoverageSummary displays the coverage summary
func showCoverageSummary(coverageFile string) {
	fmt.Printf("\n%s📊 Coverage Summary:%s\n", colorYellow, colorReset)

	coverCmd := exec.Command("go", "tool", "cover", "-func", coverageFile)
	output, err := coverCmd.Output()
	if err != nil {
		return
	}

	printFilteredCoverage(string(output))
}

// printFilteredCoverage prints filtered coverage output
func printFilteredCoverage(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if shouldPrintCoverageLine(line) {
			if strings.Contains(line, "total:") {
				fmt.Printf("\n%s%s%s\n", colorCyan, line, colorReset)
			} else {
				fmt.Println(line)
			}
		}
	}
}

// shouldPrintCoverageLine determines if a coverage line should be printed
func shouldPrintCoverageLine(line string) bool {
	return strings.Contains(line, "auth_service.go") ||
		strings.Contains(line, "user_service.go") ||
		strings.Contains(line, "total:")
}

// generateHTMLReport generates an HTML coverage report
func generateHTMLReport(coverageFile string) {
	fmt.Printf("\n%s📄 Generating HTML coverage report...%s\n", colorYellow, colorReset)

	htmlFile := "coverage_auth.html"
	htmlCmd := exec.Command("go", "tool", "cover", "-html", coverageFile, "-o", htmlFile)

	if err := htmlCmd.Run(); err != nil {
		return
	}

	fmt.Printf("✅ Report saved to: %s\n", htmlFile)

	// Ask to open report on Windows
	if runtime.GOOS == "windows" {
		offerToOpenReport(htmlFile)
	}
}

// offerToOpenReport asks the user if they want to open the report in a browser
func offerToOpenReport(htmlFile string) {
	fmt.Print("\nOpen report in browser? (y/n): ")
	var response string

	if _, err := fmt.Scanln(&response); err != nil && err.Error() != "unexpected newline" {
		fmt.Printf("\n%sWarning: Error reading response: %v%s\n", colorYellow, err, colorReset)
		return
	}

	if strings.ToLower(response) == "y" {
		if err := exec.Command("cmd", "/c", "start", htmlFile).Start(); err != nil {
			printError("Failed to open HTML file in browser", err)
		}
	}
}

// printError prints an error message
func printError(message string, err error) {
	if err != nil {
		fmt.Printf("%s❌ %s: %v%s\n", colorRed, message, err, colorReset)
	} else {
		fmt.Printf("\n%s❌ %s%s\n", colorRed, message, colorReset)
	}
}

// printFinalSummary prints the final test summary
func printFinalSummary(_ time.Duration) {
	fmt.Printf("\n%s🎉 Authentication system tests completed successfully!%s\n", colorGreen, colorReset)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review the coverage report")
	fmt.Println("  2. Add more edge case tests if needed")
	fmt.Println("  3. Run 'make test-auth' for quick testing")
	fmt.Println("  4. Consider adding integration tests")
}
