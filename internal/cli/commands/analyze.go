package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/terzigolu/josepshbrain-go/internal/api"
	"github.com/terzigolu/josepshbrain-go/internal/config"
	"github.com/urfave/cli/v2"
)

// AnalysisResult represents the analysis result of a project
type AnalysisResult struct {
	ProjectPath     string            `json:"project_path"`
	TotalFiles      int               `json:"total_files"`
	TotalDirs       int               `json:"total_dirs"`
	TotalSize       int64             `json:"total_size_bytes"`
	Languages       map[string]int    `json:"languages"`
	FileTypes       map[string]int    `json:"file_types"`
	Technologies    []string          `json:"technologies"`
	GitInfo         *GitInfo          `json:"git_info,omitempty"`
	LinesOfCode     int               `json:"lines_of_code"`
	AnalyzedAt      time.Time         `json:"analyzed_at"`
	ProjectType     string            `json:"project_type"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	MainLanguage    string            `json:"main_language"`
	LanguagePercent map[string]float64 `json:"language_percent"`
}

// GitInfo represents git repository information
type GitInfo struct {
	IsGitRepo     bool   `json:"is_git_repo"`
	CurrentBranch string `json:"current_branch,omitempty"`
	CommitCount   int    `json:"commit_count"`
	LastCommit    string `json:"last_commit,omitempty"`
}

// NewAnalyzeCommand creates the analyze command
func NewAnalyzeCommand() *cli.Command {
	return &cli.Command{
		Name:      "analyze",
		Usage:     "Analyze project structure and save to project configuration",
		ArgsUsage: "[path]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "save",
				Aliases: []string{"s"},
				Usage:   "Save analysis results to active project configuration",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format (pretty|json)",
				Value:   "pretty",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed analysis information",
			},
		},
		Action: func(c *cli.Context) error {
			// Determine project path
			projectPath := "."
			if c.NArg() > 0 {
				projectPath = c.Args().First()
			}

			// Get absolute path
			absPath, err := filepath.Abs(projectPath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %w", err)
			}

			// Check if path exists
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", absPath)
			}

			// Perform analysis
			result, err := analyzeProject(absPath, c.Bool("verbose"))
			if err != nil {
				return fmt.Errorf("failed to analyze project: %w", err)
			}

			// Output results
			if c.String("format") == "json" {
				return outputJSON(result)
			}

			outputPretty(result, c.Bool("verbose"))

			// Save to project configuration if requested
			if c.Bool("save") {
				if err := saveAnalysisToProject(result); err != nil {
					fmt.Printf("⚠️  Warning: Failed to save analysis to project: %v\n", err)
				} else {
					fmt.Printf("✅ Analysis saved to active project configuration\n")
				}
			}

			return nil
		},
	}
}

// analyzeProject performs the actual project analysis
func analyzeProject(projectPath string, verbose bool) (*AnalysisResult, error) {
	result := &AnalysisResult{
		ProjectPath:     projectPath,
		Languages:       make(map[string]int),
		FileTypes:       make(map[string]int),
		Technologies:    make([]string, 0),
		Dependencies:    make(map[string]string),
		LanguagePercent: make(map[string]float64),
		AnalyzedAt:      time.Now(),
	}

	if verbose {
		fmt.Printf("🔍 Starting analysis of: %s\n", projectPath)
	}

	// Walk through directory structure
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common build/dependency directories
		skipDirs := []string{"node_modules", "vendor", "target", "build", "dist", ".git", "__pycache__"}
		for _, skipDir := range skipDirs {
			if info.IsDir() && info.Name() == skipDir {
				return filepath.SkipDir
			}
		}

		if info.IsDir() {
			result.TotalDirs++
		} else {
			result.TotalFiles++
			result.TotalSize += info.Size()

			// Analyze file extension
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext != "" {
				result.FileTypes[ext]++
			}

			// Detect programming language
			lang := detectLanguage(info.Name(), ext)
			if lang != "" {
				result.Languages[lang]++
				
				// Count lines of code for programming files
				if isProgrammingFile(ext) {
					lines := countLines(path)
					result.LinesOfCode += lines
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Calculate language percentages
	totalLangFiles := 0
	for _, count := range result.Languages {
		totalLangFiles += count
	}
	
	if totalLangFiles > 0 {
		for lang, count := range result.Languages {
			result.LanguagePercent[lang] = float64(count) / float64(totalLangFiles) * 100
		}
	}

	// Determine main language
	maxCount := 0
	for lang, count := range result.Languages {
		if count > maxCount {
			maxCount = count
			result.MainLanguage = lang
		}
	}

	// Detect technologies and project type
	result.Technologies = detectTechnologies(projectPath)
	result.ProjectType = detectProjectType(projectPath, result.Technologies)

	// Analyze dependencies
	result.Dependencies = analyzeDependencies(projectPath)

	// Get git information
	result.GitInfo = getGitInfo(projectPath)

	return result, nil
}

// detectLanguage detects programming language from file name and extension
func detectLanguage(filename, ext string) string {
	langMap := map[string]string{
		".go":     "Go",
		".js":     "JavaScript",
		".ts":     "TypeScript",
		".jsx":    "JavaScript",
		".tsx":    "TypeScript",
		".py":     "Python",
		".java":   "Java",
		".c":      "C",
		".cpp":    "C++",
		".cc":     "C++",
		".cxx":    "C++",
		".cs":     "C#",
		".php":    "PHP",
		".rb":     "Ruby",
		".rs":     "Rust",
		".swift":  "Swift",
		".kt":     "Kotlin",
		".scala":  "Scala",
		".r":      "R",
		".sh":     "Shell",
		".bash":   "Shell",
		".zsh":    "Shell",
		".fish":   "Shell",
		".ps1":    "PowerShell",
		".html":   "HTML",
		".css":    "CSS",
		".scss":   "SCSS",
		".sass":   "SASS",
		".less":   "LESS",
		".xml":    "XML",
		".json":   "JSON",
		".yaml":   "YAML",
		".yml":    "YAML",
		".toml":   "TOML",
		".ini":    "INI",
		".cfg":    "Config",
		".conf":   "Config",
		".sql":    "SQL",
		".md":     "Markdown",
		".txt":    "Text",
		".dockerfile": "Docker",
	}

	if filename == "Dockerfile" {
		return "Docker"
	}

	if lang, exists := langMap[ext]; exists {
		return lang
	}

	return ""
}

// isProgrammingFile checks if file extension represents a programming file
func isProgrammingFile(ext string) bool {
	programmingExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".py": true, ".java": true, ".c": true, ".cpp": true, ".cc": true,
		".cxx": true, ".cs": true, ".php": true, ".rb": true, ".rs": true,
		".swift": true, ".kt": true, ".scala": true, ".r": true,
		".sh": true, ".bash": true, ".zsh": true, ".fish": true, ".ps1": true,
	}
	return programmingExts[ext]
}

// countLines counts lines in a file
func countLines(filename string) int {
	content, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}
	return strings.Count(string(content), "\n") + 1
}

// detectTechnologies detects technologies used in the project
func detectTechnologies(projectPath string) []string {
	var technologies []string
	techMap := make(map[string]bool)

	// Check for specific files that indicate technologies
	techFiles := map[string]string{
		"package.json":    "Node.js",
		"go.mod":          "Go",
		"requirements.txt": "Python",
		"Pipfile":         "Python",
		"pyproject.toml":  "Python",
		"pom.xml":         "Maven",
		"build.gradle":    "Gradle",
		"Cargo.toml":      "Rust",
		"composer.json":   "PHP",
		"Gemfile":         "Ruby",
		"Dockerfile":      "Docker",
		"docker-compose.yml": "Docker Compose",
		"Makefile":        "Make",
		"CMakeLists.txt":  "CMake",
		".gitignore":      "Git",
		"tsconfig.json":   "TypeScript",
		"webpack.config.js": "Webpack",
		"vite.config.js":  "Vite",
		"next.config.js":  "Next.js",
		"nuxt.config.js":  "Nuxt.js",
		"angular.json":    "Angular",
		"vue.config.js":   "Vue.js",
		"svelte.config.js": "Svelte",
	}

	for filename, tech := range techFiles {
		if _, err := os.Stat(filepath.Join(projectPath, filename)); err == nil {
			techMap[tech] = true
		}
	}

	// Check for framework-specific patterns
	if _, err := os.Stat(filepath.Join(projectPath, "src", "main", "java")); err == nil {
		techMap["Java"] = true
		techMap["Spring"] = true // Assume Spring if it's a Java project with src/main/java
	}

	if _, err := os.Stat(filepath.Join(projectPath, "public", "index.html")); err == nil {
		techMap["Web"] = true
	}

	// Convert map to slice
	for tech := range techMap {
		technologies = append(technologies, tech)
	}

	sort.Strings(technologies)
	return technologies
}

// detectProjectType determines the project type based on analysis
func detectProjectType(projectPath string, technologies []string) string {
	techSet := make(map[string]bool)
	for _, tech := range technologies {
		techSet[tech] = true
	}

	// Check for specific project types
	if techSet["Go"] {
		return "Go Application"
	}
	if techSet["Node.js"] {
		if techSet["Next.js"] {
			return "Next.js Application"
		}
		if techSet["Vue.js"] {
			return "Vue.js Application"
		}
		if techSet["Angular"] {
			return "Angular Application"
		}
		if techSet["Svelte"] {
			return "Svelte Application"
		}
		return "Node.js Application"
	}
	if techSet["Python"] {
		return "Python Application"
	}
	if techSet["Java"] {
		return "Java Application"
	}
	if techSet["Rust"] {
		return "Rust Application"
	}
	if techSet["PHP"] {
		return "PHP Application"
	}
	if techSet["Ruby"] {
		return "Ruby Application"
	}
	if techSet["Docker"] {
		return "Containerized Application"
	}
	if techSet["Web"] {
		return "Web Application"
	}

	return "Unknown"
}

// analyzeDependencies analyzes project dependencies
func analyzeDependencies(projectPath string) map[string]string {
	dependencies := make(map[string]string)

	// Go dependencies
	if content, err := os.ReadFile(filepath.Join(projectPath, "go.mod")); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "go ") {
				dependencies["go_version"] = strings.TrimPrefix(line, "go ")
			}
		}
	}

	// Node.js dependencies
	if content, err := os.ReadFile(filepath.Join(projectPath, "package.json")); err == nil {
		var packageData map[string]interface{}
		if json.Unmarshal(content, &packageData) == nil {
			if deps, ok := packageData["dependencies"].(map[string]interface{}); ok {
				dependencies["node_deps"] = fmt.Sprintf("%d", len(deps))
			}
			if devDeps, ok := packageData["devDependencies"].(map[string]interface{}); ok {
				dependencies["node_dev_deps"] = fmt.Sprintf("%d", len(devDeps))
			}
		}
	}

	// Python dependencies
	if content, err := os.ReadFile(filepath.Join(projectPath, "requirements.txt")); err == nil {
		lines := strings.Split(string(content), "\n")
		count := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
				count++
			}
		}
		dependencies["python_deps"] = fmt.Sprintf("%d", count)
	}

	return dependencies
}

// getGitInfo gets git repository information
func getGitInfo(projectPath string) *GitInfo {
	gitDir := filepath.Join(projectPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return &GitInfo{IsGitRepo: false}
	}

	gitInfo := &GitInfo{IsGitRepo: true}

	// Try to get current branch
	if headContent, err := os.ReadFile(filepath.Join(gitDir, "HEAD")); err == nil {
		headStr := strings.TrimSpace(string(headContent))
		if strings.HasPrefix(headStr, "ref: refs/heads/") {
			gitInfo.CurrentBranch = strings.TrimPrefix(headStr, "ref: refs/heads/")
		}
	}

	return gitInfo
}

// outputJSON outputs analysis results in JSON format
func outputJSON(result *AnalysisResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputPretty outputs analysis results in a pretty format
func outputPretty(result *AnalysisResult, verbose bool) {
	fmt.Printf("📊 Project Analysis: %s\n", filepath.Base(result.ProjectPath))
	fmt.Printf("================================\n")
	fmt.Printf("📂 Structure: %d dirs, %d files\n", result.TotalDirs, result.TotalFiles)
	fmt.Printf("💾 Size: %s\n", formatBytes(result.TotalSize))
	fmt.Printf("🔧 Type: %s\n", result.ProjectType)
	fmt.Printf("📝 Main Language: %s\n", result.MainLanguage)
	fmt.Printf("📈 Lines of Code: %s\n", formatNumber(result.LinesOfCode))

	if len(result.Technologies) > 0 {
		fmt.Printf("🛠️  Technologies: %s\n", strings.Join(result.Technologies, ", "))
	}

	if len(result.Languages) > 0 {
		fmt.Printf("🌐 Languages:\n")
		// Sort languages by count
		type langCount struct {
			lang  string
			count int
			percent float64
		}
		var langs []langCount
		for lang, count := range result.Languages {
			langs = append(langs, langCount{lang, count, result.LanguagePercent[lang]})
		}
		sort.Slice(langs, func(i, j int) bool {
			return langs[i].count > langs[j].count
		})
		
		for _, lc := range langs {
			fmt.Printf("   %s: %d files (%.1f%%)\n", lc.lang, lc.count, lc.percent)
		}
	}

	if result.GitInfo != nil && result.GitInfo.IsGitRepo {
		fmt.Printf("🌿 Git: %s branch\n", result.GitInfo.CurrentBranch)
	}

	if verbose && len(result.Dependencies) > 0 {
		fmt.Printf("\n📦 Dependencies:\n")
		for key, value := range result.Dependencies {
			fmt.Printf("   %s: %s\n", key, value)
		}
	}

	if verbose && len(result.FileTypes) > 0 {
		fmt.Printf("\n📄 File Types:\n")
		// Sort file types by count
		type fileTypeCount struct {
			ext   string
			count int
		}
		var fileTypes []fileTypeCount
		for ext, count := range result.FileTypes {
			fileTypes = append(fileTypes, fileTypeCount{ext, count})
		}
		sort.Slice(fileTypes, func(i, j int) bool {
			return fileTypes[i].count > fileTypes[j].count
		})
		
		for _, ft := range fileTypes {
			fmt.Printf("   %s: %d files\n", ft.ext, ft.count)
		}
	}

	fmt.Printf("\n🕐 Analyzed at: %s\n", result.AnalyzedAt.Format("2006-01-02 15:04:05"))
}

// saveAnalysisToProject saves analysis results to the active project configuration
func saveAnalysisToProject(result *AnalysisResult) error {
	// Load config to get active project
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.ActiveProjectID == "" {
		return fmt.Errorf("no active project set. Use 'jbraincli project use <id>' to set an active project")
	}

	// Prepare analysis data for project configuration
	analysisData := map[string]interface{}{
		"analysis": result,
		"path":     result.ProjectPath,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(analysisData)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis data: %w", err)
	}

	// Update project with analysis results
	client := api.NewClient()
	updateData := map[string]interface{}{
		"configuration": json.RawMessage(jsonData),
		"path":          result.ProjectPath,
	}

	_, err = client.UpdateProject(cfg.ActiveProjectID, updateData)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// formatBytes formats bytes to human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatNumber formats number with thousand separators
func formatNumber(num int) string {
	str := strconv.Itoa(num)
	n := len(str)
	if n <= 3 {
		return str
	}
	
	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}
	return result.String()
}