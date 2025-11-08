package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// BuildTarget represents a buildable binary target
type BuildTarget struct {
	Name        string
	SourcePath  string
	BinaryName  string
	Description string
}

// Available build targets
var buildTargets = map[string]BuildTarget{
	"gameTaskEmulator": {
		Name:        "gameTaskEmulator",
		SourcePath:  "./cmd/gameTaskEmulator",
		BinaryName:  "gameTaskEmulator",
		Description: "NHL game tracker scheduler that creates Cloud Tasks for game monitoring",
	},
}

func main() {
	var (
		target = flag.String("target", "", "Target to build (gameTaskEmulator)")
		list   = flag.Bool("list", false, "List available build targets")
		all    = flag.Bool("all", false, "Build all available targets")
	)
	flag.Parse()

	// Show usage if no flags provided
	if len(os.Args) == 1 {
		showUsage()
		return
	}

	// List available targets
	if *list {
		listTargets()
		return
	}

	// Create bin directory if it doesn't exist
	binDir := "./bin"
	if err := os.MkdirAll(binDir, 0755); err != nil {
		log.Fatalf("Failed to create bin directory: %v", err)
	}

	// Build all targets
	if *all {
		buildAllTargets(binDir)
		return
	}

	// Build specific target
	if *target == "" {
		fmt.Println("Error: target flag is required")
		showUsage()
		os.Exit(1)
	}

	buildTarget(*target, binDir)
}

func showUsage() {
	fmt.Println("Build system for gameTaskEmulator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run build.go -target <target>  Build specific target")
	fmt.Println("  go run build.go -all              Build all targets")
	fmt.Println("  go run build.go -list             List available targets")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run build.go -target gameTaskEmulator")
	fmt.Println("  go run build.go -all")
	fmt.Println()
	fmt.Println("All binaries are saved to ./bin/")
}

func listTargets() {
	fmt.Println("Available build targets:")
	fmt.Println()
	for name, target := range buildTargets {
		fmt.Printf("  %-25s %s\n", name, target.Description)
		fmt.Printf("  %-25s Source: %s\n", "", target.SourcePath)
		fmt.Printf("  %-25s Binary: ./bin/%s\n", "", target.BinaryName)
		fmt.Println()
	}
}

func buildAllTargets(binDir string) {
	fmt.Println("Building all targets...")
	fmt.Println()

	success := 0
	total := len(buildTargets)

	for name := range buildTargets {
		if buildTarget(name, binDir) {
			success++
		}
	}

	fmt.Println()
	fmt.Printf("Build complete: %d/%d targets built successfully\n", success, total)
}

func buildTarget(targetName, binDir string) bool {
	target, exists := buildTargets[targetName]
	if !exists {
		fmt.Printf("Error: Unknown target '%s'\n", targetName)
		fmt.Println("Use -list to see available targets")
		return false
	}

	fmt.Printf("Building %s...\n", target.Name)

	// Check if source directory exists
	if _, err := os.Stat(target.SourcePath); os.IsNotExist(err) {
		fmt.Printf("Error: Source path does not exist: %s\n", target.SourcePath)
		return false
	}

	// Get absolute paths
	absOutputPath, err := filepath.Abs(filepath.Join(binDir, target.BinaryName))
	if err != nil {
		fmt.Printf("Error getting absolute output path: %v\n", err)
		return false
	}

	absSourcePath, err := filepath.Abs(target.SourcePath)
	if err != nil {
		fmt.Printf("Error getting absolute source path: %v\n", err)
		return false
	}

	// For standard cmd structure - go up two levels from cmd/binary to module root
	moduleDir := filepath.Dir(filepath.Dir(absSourcePath))
	relativePath := "./" + filepath.Join("cmd", target.BinaryName)

	cmd := exec.Command("go", "build", "-o", absOutputPath, relativePath)

	// Set working directory to the module directory
	cmd.Dir = moduleDir

	// Set environment variables for local builds
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
	)

	// Run the build
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error building %s: %v\n", target.Name, err)
		if len(output) > 0 {
			fmt.Printf("Build output:\n%s\n", string(output))
		}
		return false
	}

	// Check if binary was created
	if _, err := os.Stat(absOutputPath); err != nil {
		fmt.Printf("Error: Binary was not created at %s\n", absOutputPath)
		return false
	}

	fmt.Printf("✓ Successfully built %s -> %s\n", target.Name, absOutputPath)
	return true
}
