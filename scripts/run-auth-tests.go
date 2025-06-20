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

func main() {
	fmt.Printf("%s🔐 ASAM Backend - Auth System Test Runner%s\n", colorCyan, colorReset)
	fmt.Println(strings.Repeat("=", 50))

	// Get project root
	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Printf("%s❌ Error getting current directory: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	// Change to project root if we're in scripts directory
	if strings.HasSuffix(projectRoot, "scripts") {
		projectRoot = filepath.Dir(projectRoot)
		os.Chdir(projectRoot)
	}

	// Test package path
	testPath := "./test/internal/domain/services"
	coverageFile := "coverage_auth.out"
	coveragePackage := "github.com/javicabdev/asam-backend/internal/domain/services"

	// Step 1: Check if test files exist
	fmt.Printf("\n%s📁 Checking test files...%s\n", colorYellow, colorReset)
	testFiles := []string{
		"test/internal/domain/services/auth_service_test.go",
		"test/internal/domain/services/user_service_test.go",
	}

	allFilesExist := true
	for _, file := range testFiles {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("  ✅ %s\n", file)
		} else {
			fmt.Printf("  ❌ %s (not found)\n", file)
			allFilesExist = false
		}
	}

	if !allFilesExist {
		fmt.Printf("\n%s❌ Some test files are missing!%s\n", colorRed, colorReset)
		os.Exit(1)
	}

	// Step 2: Run tests
	fmt.Printf("\n%s🧪 Running authentication tests...%s\n", colorYellow, colorReset)

	start := time.Now()

	// Build test command
	cmd := exec.Command("go", "test", "-v", "-race",
		"-coverprofile="+coverageFile,
		"-covermode=atomic",
		"-coverpkg="+coveragePackage,
		testPath)

	// Set output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run tests
	err = cmd.Run()
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("\n%s❌ Tests failed! %s\n", colorRed, colorReset)
		os.Exit(1)
	}

	fmt.Printf("\n%s✅ All tests passed in %.2fs!%s\n", colorGreen, duration.Seconds(), colorReset)

	// Step 3: Show coverage summary
	fmt.Printf("\n%s📊 Coverage Summary:%s\n", colorYellow, colorReset)

	coverCmd := exec.Command("go", "tool", "cover", "-func="+coverageFile)
	output, err := coverCmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "auth_service.go") ||
				strings.Contains(line, "user_service.go") ||
				strings.Contains(line, "total:") {
				if strings.Contains(line, "total:") {
					fmt.Printf("\n%s%s%s\n", colorCyan, line, colorReset)
				} else {
					fmt.Println(line)
				}
			}
		}
	}

	// Step 4: Generate HTML report
	fmt.Printf("\n%s📄 Generating HTML coverage report...%s\n", colorYellow, colorReset)

	htmlFile := "coverage_auth.html"
	htmlCmd := exec.Command("go", "tool", "cover", "-html="+coverageFile, "-o", htmlFile)
	err = htmlCmd.Run()

	if err == nil {
		fmt.Printf("✅ Report saved to: %s\n", htmlFile)

		// Ask to open report
		if runtime.GOOS == "windows" {
			fmt.Print("\nOpen report in browser? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) == "y" {
				exec.Command("cmd", "/c", "start", htmlFile).Start()
			}
		}
	}

	// Final summary
	fmt.Printf("\n%s🎉 Authentication system tests completed successfully!%s\n", colorGreen, colorReset)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review the coverage report")
	fmt.Println("  2. Add more edge case tests if needed")
	fmt.Println("  3. Run 'make test-auth' for quick testing")
	fmt.Println("  4. Consider adding integration tests")
}
