package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// Template for the CMakeLists.txt
const cmakeTemplate = `
cmake_minimum_required(VERSION {{.CMakeVersion}})

# Set the project name and language
project({{.ProjectName}} LANGUAGES {{.CMakeLang}})

# Set the project language standard
{{if eq .Lang "cpp"}}
set(CMAKE_CXX_STANDARD {{.Standard}})
{{else if eq .Lang "c"}}
set(CMAKE_C_STANDARD {{.Standard}})
{{end}}

# Add the executable or library
{{if eq .Type "executable"}}
add_executable({{.ProjectName}} src/main.{{.FileExt}})
{{else if eq .Type "library"}}
add_library({{.ProjectName}} src/main.{{.FileExt}})
{{end}}
`

// Template for a basic main.cpp file
const cppMainTemplate = `
#include <iostream>

int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}
`

// .gitignore template for C and C++ projects
const gitignoreTemplate = `
# Compiled Object files
*.o
*.obj

# Precompiled Headers
*.gch
*.pch

# Compiled Dynamic libraries
*.so
*.dylib
*.dll

# Compiled Static libraries
*.lib
*.a

# Executable files
*.exe
*.out
*.app

# CMake Build
/build/
CMakeCache.txt
CMakeFiles/
cmake_install.cmake
Makefile

# IDE and editor files
.vscode/
*.swp
*.swo
*.idea/
*.vscode/
`

// .gitattributes template for consistency in line endings
const gitattributesTemplate = `
# Ensure consistent line endings
* text=auto

# Treat C and C++ files as text
*.c text
*.cpp text
*.h text
*.hpp text
`

// ProjectData holds information about the project
type ProjectData struct {
	ProjectName  string
	Type         string
	Lang         string
	Standard     string
	CMakeVersion string
	CMakeLang    string
	FileExt      string
}

// Helper function to retrieve the installed CMake version
func getCMakeVersion() (string, error) {
	cmd := exec.Command("cmake", "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`cmake version (\d+\.\d+\.\d+)`)
	match := re.FindStringSubmatch(string(out))
	if len(match) < 2 {
		return "", fmt.Errorf("unable to parse CMake version")
	}

	return match[1], nil
}

func main() {
	// Define command-line flags
	projectName := flag.String("name", "", "Name of the project")
	projectType := flag.String("type", "executable", "Project type (executable, library)")
	lang := flag.String("lang", "cpp", "Programming language (cpp, c)")
	standard := flag.String("std", "11", "Language standard (e.g., 11, 14, 17 for C++)")

	// Parse flags
	flag.Parse()

	if *projectName == "" {
		fmt.Println("Error: Project name is required. Use -name to specify the project name.")
		os.Exit(1)
	}

	if *lang != "cpp" && *lang != "c" {
		fmt.Println("Error: Unsupported language. Supported options: cpp, c")
		os.Exit(1)
	}

	// Get the installed CMake version
	cmakeVersion, err := getCMakeVersion()
	if err != nil {
		fmt.Printf("Error retrieving CMake version: %v\n", err)
		os.Exit(1)
	}

	// Set CMakeLang and file extension based on language
	var cmakeLang, fileExt string
	if *lang == "cpp" {
		cmakeLang = "CXX"
		fileExt = "cpp"
	} else if *lang == "c" {
		cmakeLang = "C"
		fileExt = "c"
	}

	// Set up the project data
	projectData := ProjectData{
		ProjectName:  *projectName,
		Type:         *projectType,
		Lang:         *lang,
		Standard:     *standard,
		CMakeVersion: cmakeVersion,
		CMakeLang:    cmakeLang,
		FileExt:      fileExt,
	}

	// Create project structure
	createProjectStructure(projectData)

	// Prompt the user for version control initialization
	initializeVersionControl(projectData.ProjectName)
}

// Function to create the project structure
func createProjectStructure(data ProjectData) {
	dirs := []string{
		"src",
		"include",
		"build",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(data.ProjectName, dir), os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Generate CMakeLists.txt
	createCMakeLists(data)

	// Create a basic main file depending on the language
	if data.Lang == "cpp" {
		createFile(filepath.Join(data.ProjectName, "src", "main.cpp"), cppMainTemplate)
	} else if data.Lang == "c" {
		createFile(filepath.Join(data.ProjectName, "src", "main.c"), cppMainTemplate)
	}

	fmt.Printf("Project %s created successfully.\n", data.ProjectName)
}

// Function to generate the CMakeLists.txt file
func createCMakeLists(data ProjectData) {
	tmpl, err := template.New("cmake").Parse(cmakeTemplate)
	if err != nil {
		fmt.Printf("Error creating CMakeLists.txt template: %v\n", err)
		os.Exit(1)
	}

	file, err := os.Create(filepath.Join(data.ProjectName, "CMakeLists.txt"))
	if err != nil {
		fmt.Printf("Error creating CMakeLists.txt file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		fmt.Printf("Error writing to CMakeLists.txt: %v\n", err)
		os.Exit(1)
	}
}

// Function to create a file and write contents to it
func createFile(path, content string) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", path, err)
		os.Exit(1)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Printf("Error writing to file %s: %v\n", path, err)
		os.Exit(1)
	}
}

// Function to prompt the user to initialize Git and execute git init
func initializeVersionControl(projectPath string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to initialize Git version control? (y/n): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		// Run 'git init' in the project directory
		cmd := exec.Command("git", "init")
		cmd.Dir = projectPath // Set working directory to the project path
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error initializing Git repository: %v\n", err)
			return
		}

		// Create .gitignore and .gitattributes
		createFile(filepath.Join(projectPath, ".gitignore"), gitignoreTemplate)
		createFile(filepath.Join(projectPath, ".gitattributes"), gitattributesTemplate)

		// Automatically add the files to Git
		cmdAdd := exec.Command("git", "add", ".gitignore", ".gitattributes")
		cmdAdd.Dir = projectPath
		err = cmdAdd.Run()
		if err != nil {
			fmt.Printf("Error adding .gitignore/.gitattributes to Git: %v\n", err)
			return
		}

		fmt.Println("Git repository initialized successfully with .gitignore and .gitattributes.")
	} else {
		fmt.Println("Skipped Git initialization.")
	}
}
