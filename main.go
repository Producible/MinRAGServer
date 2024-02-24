package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type GeneralSettings struct {
	ServerPort                     string `json:"server_port"`
	DisableExternalNetworkBrowsing bool   `json:"disable_external_network_browsing"`
	ShowHidden                     bool   `json:"show_hidden"`
	TimeStamp                      bool   `json:"time_stamp"`
	InclusiveExtensions            string `json:"inclusive_extensions"`
	ExclusiveExtensions            string `json:"exclusive_extensions"`
	ExclusiveFolders               string `json:"exclusive_folders"`
}

var generalSettings GeneralSettings

type Config struct {
	ProjectName         string `json:"project_name"`
	RootPath            string `json:"root_path"`
	ProjectURL          string `json:"project_url"`
	InclusiveExtensions string `json:"inclusive_extensions,omitempty"`
	ExclusiveExtensions string `json:"exclusive_extensions,omitempty"`
	ExclusiveFolders    string `json:"exclusive_folders,omitempty"`
	ExclusiveFiles      string `json:"exclusive_files,omitempty"`
}

var configs map[string]Config
var selectedConfig Config

// Helper function to check if an IP address is local
func isLocalIP(ip string) bool {
	localIPBlocks := []string{
		"127.0.0.1/8",    // IPv4 loopback
		"::1/128",        // IPv6 loopback
		"10.0.0.0/8",     // Private-Use Networks
		"172.16.0.0/12",  // Private-Use Networks
		"192.168.0.0/16", // Private-Use Networks
	}

	for _, block := range localIPBlocks {
		_, cidr, _ := net.ParseCIDR(block)
		if cidr.Contains(net.ParseIP(ip)) {
			return true
		}
	}
	return false
}

func loadConfigs() error {
	// Load general settings
	settingsData, err := os.ReadFile("settings.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(settingsData, &generalSettings)
	if err != nil {
		return err
	}

	configs = make(map[string]Config)
	files, err := os.ReadDir("config")
	if err != nil {
		return err
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			configPath := filepath.Join("config", file.Name())
			configFile, err := os.Open(configPath)
			if err != nil {
				return err
			}
			defer configFile.Close()

			var config Config
			decoder := json.NewDecoder(configFile)
			err = decoder.Decode(&config)
			if err != nil {
				return err
			}

			configs[file.Name()] = config
		}
	}
	return nil
}

func projectHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request is from a local IP
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if generalSettings.DisableExternalNetworkBrowsing && !isLocalIP(ip) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	project := vars["project_json_name"]
	if project == "" || configs[project+".json"].ProjectName == "" {
		for filename, config := range configs {
			projectID := strings.TrimSuffix(filename, ".json")
			fmt.Fprintf(w, "<a href='/p/%s'>%s</a><br>", projectID, config.ProjectName)
		}
		return
	}

	selectedConfig = configs[project+".json"]
	dirPath := selectedConfig.RootPath

	// Generate links for the root directory
	dirStructureLink := fmt.Sprintf("/s/%s/", project)
	dirContentsLink := fmt.Sprintf("/c/%s/", project)
	dirStructureUrl := fmt.Sprintf("%s%s", selectedConfig.ProjectURL, dirStructureLink)
	dirContentsUrl := fmt.Sprintf("%s%s", selectedConfig.ProjectURL, dirContentsLink)

	fmt.Fprintln(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>MinRAGServer</title>
<link rel="stylesheet" href="/static/style.css">
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.3/css/all.min.css">
<script>
    var appendTimestamp = `+fmt.Sprintf("%t", generalSettings.TimeStamp)+`;
</script>
</head>
<body>
<a href="/" class="back-button"><i class="fas fa-arrow-left"></i> Projects</a>
<h1>`+selectedConfig.ProjectName+`</h1>
<div class="tree-view">
        <ul>
            <li class="root-item expanded">
                <div class='item'>
                    <span>`+selectedConfig.ProjectName+`</span>
                    <a href='`+dirStructureLink+`' target='_blank' class='buttons' title="Display whole structure"><i class='fas fa-sitemap' style='color:orange'></i></a>
                    <a href='`+dirContentsLink+`' target='_blank' class='buttons' title="Display all file content"><i class='fas fa-file-code' style='color:#6495ED'></i></a>
                    <button class='copy-button buttons' data-url='`+dirStructureUrl+`' title="Copy structure URL"><i class='fas fa-copy' style='color:#20B2AA'></i></button>
                    <button class='copy-button buttons' data-url='`+dirContentsUrl+`' title="Copy file-content URL"><i class='fas fa-copy' style='color:green'></i></button>
                </div>
                <ul>`)

	writeDirectory(w, dirPath, dirPath, project)

	fmt.Fprintln(w, `</ul></li>
        </ul>
    </div>
<script src="/static/script.js"></script>
<script src="/static/clipboard.js"></script>
</body>
</html>`)
}

func fileViewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project_json_name"]
	path := vars["relativePath"]

	if project == "" || configs[project+".json"].ProjectName == "" {
		http.Error(w, "Invalid project", http.StatusBadRequest)
		return
	}

	selectedConfig = configs[project+".json"]
	fullPath := filepath.Join(selectedConfig.RootPath, path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>File Content</title>
</head>
<body>
<pre>`+string(data)+`</pre>
</body>
</html>`)
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project_json_name"]
	path := vars["relativePath"]

	if project == "" || configs[project+".json"].ProjectName == "" {
		http.Error(w, "Invalid project", http.StatusBadRequest)
		return
	}

	selectedConfig = configs[project+".json"]
	fullPath := filepath.Join(selectedConfig.RootPath, path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the content type to plain text with UTF-8 charset
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	// Write the file content as plain text
	w.Write(data)
}

func jsonFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project_json_name"]
	path := vars["relativePath"]

	if project == "" || configs[project+".json"].ProjectName == "" {
		http.Error(w, "Invalid project", http.StatusBadRequest)
		return
	}

	selectedConfig = configs[project+".json"]
	fullPath := filepath.Join(selectedConfig.RootPath, path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(data), "\n")
	jsonData := make([]map[string]interface{}, len(lines))
	for i, line := range lines {
		jsonData[i] = map[string]interface{}{
			"line":    i + 1,
			"content": line,
		}
	}

	fileName := filepath.Base(fullPath)          // Get only the file name
	relativePath := "/" + filepath.ToSlash(path) // Ensure path starts with "/"

	response := map[string]interface{}{
		"file":                     fileName,
		"path":                     relativePath,
		"content-with-line-number": jsonData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func dirStructureHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project_json_name"]
	path := vars["relativePath"]

	if project == "" || configs[project+".json"].ProjectName == "" {
		http.Error(w, "Invalid project", http.StatusBadRequest)
		return
	}

	selectedConfig = configs[project+".json"]
	fullPath := filepath.Join(selectedConfig.RootPath, path)

	// Get the configurations from the selected project or from general settings
	inclusiveExtensions := strings.Split(selectedConfig.InclusiveExtensions, ",")
	if inclusiveExtensions[0] == "" {
		inclusiveExtensions = strings.Split(generalSettings.InclusiveExtensions, ",")
	}
	exclusiveExtensions := strings.Split(selectedConfig.ExclusiveExtensions, ",")
	if exclusiveExtensions[0] == "" {
		exclusiveExtensions = strings.Split(generalSettings.ExclusiveExtensions, ",")
	}
	exclusiveFolders := strings.Split(selectedConfig.ExclusiveFolders, ",")
	if exclusiveFolders[0] == "" {
		exclusiveFolders = strings.Split(generalSettings.ExclusiveFolders, ",")
	}
	exclusiveFiles := strings.Split(selectedConfig.ExclusiveFiles, ",")

	var buildDirStructure func(string, string, int) string
	buildDirStructure = func(currentPath, relativePath string, level int) string {
		files, err := os.ReadDir(currentPath)
		if err != nil {
			return "Error reading directory: " + err.Error() + "\n"
		}

		var structure string
		indent := strings.Repeat("  ", level)
		for _, file := range files {
			if !generalSettings.ShowHidden && strings.HasPrefix(file.Name(), ".") {
				continue // Skip hidden files and directories
			}
			fileName := file.Name()
			if contains(exclusiveFiles, fileName) {
				continue // Skip exclusive files
			}
			if file.IsDir() {
				dirRelativePath := filepath.Join(relativePath, fileName)
				dirRelativePath = filepath.ToSlash(filepath.Clean(dirRelativePath))
				fileRelativePathToCompare := dirRelativePath
				if !strings.HasPrefix(fileRelativePathToCompare, "/") {
					fileRelativePathToCompare = fmt.Sprintf("/%s", fileRelativePathToCompare)
				}
				if checkExclusiveDir(exclusiveFolders, fileRelativePathToCompare, fileName) {
					continue
				}
				structure += fmt.Sprintf("%s[/%s]\n", indent, dirRelativePath)
				structure += buildDirStructure(filepath.Join(currentPath, fileName), dirRelativePath, level+1)
			} else {
				ext := filepath.Ext(fileName)
				if len(ext) > 0 {
					ext = ext[1:] // Remove the leading "."
				}
				if (len(inclusiveExtensions) == 0 || inclusiveExtensions[0] == "*" || contains(inclusiveExtensions, ext)) &&
					(len(exclusiveExtensions) == 0 || !contains(exclusiveExtensions, ext)) {
					fileRelativePath := filepath.Join(relativePath, fileName)
					fileRelativePath = filepath.ToSlash(fileRelativePath)
					structure += fmt.Sprintf("%s/%s\n", indent, fileRelativePath)
				}
			}
		}
		return structure
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte(buildDirStructure(fullPath, path, 0)))
}

func dirContentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project_json_name"]
	path := vars["relativePath"]

	if project == "" || configs[project+".json"].ProjectName == "" {
		http.Error(w, "Invalid project", http.StatusBadRequest)
		return
	}

	selectedConfig = configs[project+".json"]
	fullPath := filepath.Join(selectedConfig.RootPath, path)

	// Get the configurations from the selected project or from general settings
	inclusiveExtensions := strings.Split(selectedConfig.InclusiveExtensions, ",")
	if inclusiveExtensions[0] == "" {
		inclusiveExtensions = strings.Split(generalSettings.InclusiveExtensions, ",")
	}
	exclusiveExtensions := strings.Split(selectedConfig.ExclusiveExtensions, ",")
	if exclusiveExtensions[0] == "" {
		exclusiveExtensions = strings.Split(generalSettings.ExclusiveExtensions, ",")
	}
	exclusiveFolders := strings.Split(selectedConfig.ExclusiveFolders, ",")
	if exclusiveFolders[0] == "" {
		exclusiveFolders = strings.Split(generalSettings.ExclusiveFolders, ",")
	}
	exclusiveFiles := strings.Split(selectedConfig.ExclusiveFiles, ",")

	var readDirContents func(string, string) (string, error)
	readDirContents = func(currentPath, relativePath string) (string, error) {
		var contents string
		files, err := os.ReadDir(currentPath)
		if err != nil {
			return "", err
		}

		for _, file := range files {
			if !generalSettings.ShowHidden && strings.HasPrefix(file.Name(), ".") {
				continue // Skip hidden files and directories
			}
			fileName := file.Name()
			if contains(exclusiveFiles, fileName) {
				continue // Skip exclusive files
			}
			filePath := filepath.Join(currentPath, fileName)
			fileRelativePath := filepath.Join(relativePath, fileName)
			fileRelativePath = filepath.ToSlash(filepath.Clean(fileRelativePath))
			if file.IsDir() {
				fileRelativePathToCompare := fileRelativePath
				if !strings.HasPrefix(fileRelativePathToCompare, "/") {
					fileRelativePathToCompare = fmt.Sprintf("/%s", fileRelativePathToCompare)
				}
				if checkExclusiveDir(exclusiveFolders, fileRelativePathToCompare, fileName) {
					continue
				}
				dirContents, err := readDirContents(filePath, fileRelativePath)
				if err != nil {
					return "", err
				}
				contents += dirContents
			} else {
				ext := filepath.Ext(fileName)
				if len(ext) > 0 {
					ext = ext[1:] // Remove the leading "."
				}
				if (len(inclusiveExtensions) == 0 || inclusiveExtensions[0] == "*" || contains(inclusiveExtensions, ext)) &&
					(len(exclusiveExtensions) == 0 || !contains(exclusiveExtensions, ext)) {
					fileData, err := os.ReadFile(filePath)
					if err != nil {
						return "", err
					}
					contents += fmt.Sprintf("---------------\nFile: /%s:\n\n%s\n\n", fileRelativePath, string(fileData))
				}
			}
		}
		return contents, nil
	}

	dirContents, err := readDirContents(fullPath, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte(dirContents))
}

func writeDirectory(w http.ResponseWriter, path string, rootPath string, project string) {
	files, err := os.ReadDir(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Separate directories and files
	var dirs []os.DirEntry
	var filesOnly []os.DirEntry
	for _, file := range files {
		// Use the show_hidden property from the general settings
		if !generalSettings.ShowHidden && strings.HasPrefix(file.Name(), ".") {
			continue
		}

		if file.IsDir() {
			dirs = append(dirs, file)
		} else {
			filesOnly = append(filesOnly, file)
		}
	}

	// Optionally sort dirs and filesOnly slices alphabetically
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })
	sort.Slice(filesOnly, func(i, j int) bool { return filesOnly[i].Name() < filesOnly[j].Name() })

	// Get the configurations from the selected project or from general settings
	inclusiveExtensions := strings.Split(selectedConfig.InclusiveExtensions, ",")
	if inclusiveExtensions[0] == "" {
		inclusiveExtensions = strings.Split(generalSettings.InclusiveExtensions, ",")
	}
	exclusiveExtensions := strings.Split(selectedConfig.ExclusiveExtensions, ",")
	if exclusiveExtensions[0] == "" {
		exclusiveExtensions = strings.Split(generalSettings.ExclusiveExtensions, ",")
	}
	exclusiveFolders := strings.Split(selectedConfig.ExclusiveFolders, ",")
	if exclusiveFolders[0] == "" {
		exclusiveFolders = strings.Split(generalSettings.ExclusiveFolders, ",")
	}

	exclusiveFiles := strings.Split(selectedConfig.ExclusiveFiles, ",")

	// Inside the writeDirectory function
	for _, dir := range dirs {
		relativePath := strings.TrimPrefix(path, rootPath)
		dirPath := filepath.ToSlash(filepath.Clean(filepath.Join(relativePath, dir.Name())))
		if !strings.HasPrefix(dirPath, "/") {
			dirPath = fmt.Sprintf("/%s", dirPath)
		}
		if checkExclusiveDir(exclusiveFolders, dirPath, dir.Name()) {
			continue
		}
		dirStructureLink := filepath.ToSlash(filepath.Clean(fmt.Sprintf("/s/%s%s", project, dirPath)))
		dirContentsLink := filepath.ToSlash(filepath.Clean(fmt.Sprintf("/c/%s%s", project, dirPath)))
		dirStructureUrl := filepath.ToSlash(fmt.Sprintf("%s%s", selectedConfig.ProjectURL, dirStructureLink))
		dirContentsUrl := filepath.ToSlash(fmt.Sprintf("%s%s", selectedConfig.ProjectURL, dirContentsLink))

		fmt.Fprintf(w, `<li><div class='item'><span>%s</span> 
			<a href='%s' target='_blank' class='buttons' title="Display structure in directory"><i class='fas fa-sitemap' style='color:orange'></i></a>
			<a href='%s' target='_blank' class='buttons' title="Display file content in directory"><i class='fas fa-file-code' style='color:#6495ED'></i></a>
			<button class='copy-button buttons' data-url='%s' title="Copy structure URL"><i class='fas fa-copy' style='color:#20B2AA'></i></button>
			<button class='copy-button buttons' data-url='%s' title="Copy file-content URL"><i class='fas fa-copy' style='color:green'></i></button>
		</div><ul>`, dir.Name(), dirStructureLink, dirContentsLink, dirStructureUrl, dirContentsUrl)
		writeDirectory(w, filepath.Join(path, dir.Name()), rootPath, project) // Recursive call
		fmt.Fprintln(w, "</ul></li>")
	}

	// Process files
	for _, file := range filesOnly {
		if contains(exclusiveFiles, file.Name()) {
			continue // Skip exclusive files
		}
		ext := filepath.Ext(file.Name()) // Get file extension without the leading "."

		// when ext not empty, remove "."
		if len(ext) > 0 {
			ext = ext[1:]
		}

		if (len(inclusiveExtensions) == 0 || inclusiveExtensions[0] == "*" || contains(inclusiveExtensions, ext)) &&
			(len(exclusiveExtensions) == 0 || !contains(exclusiveExtensions, ext)) {
			relativePath := strings.TrimPrefix(filepath.Join(path, file.Name()), rootPath)
			relativePath = filepath.ToSlash(relativePath)

			fileLink := fmt.Sprintf("f/%s%s", project, relativePath)
			fileViewLink := fmt.Sprintf("v/%s%s", project, relativePath)

			url := fmt.Sprintf("%s/%s", selectedConfig.ProjectURL, fileLink)
			info := fmt.Sprintf("%s: %s", file.Name(), url)
			jsonLink := fmt.Sprintf("%s/j/%s%s", selectedConfig.ProjectURL, project, relativePath)
			fmt.Fprintf(w, `<li><div class='item'>
				<a href='/%s' target='_blank' title="Display in internal URL">%s</a>
				<a href='%s' target='_blank' class='buttons' title="Display in external URL"><i class='fas fa-external-link-alt' style='color:orange'></i></a>
				<button class='copy-button buttons' data-url='%s' title="Copy external URL"><i class='fas fa-copy' style='color:#20B2AA'></i></button>
				<button class='copy-button-info buttons' data-info='%s' title="Copy external URL with path"><i class='fas fa-copy'></i></button>
				<a href='%s' target='_blank' class='buttons' title="Display content in JSON"><i class='fas fa-file-code' style='color:#87CEFA'></i></a>
			</div></li>`, fileViewLink, file.Name(), url, url, info, jsonLink)
		}
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func checkExclusiveDir(exclusiveFolders []string, relativePath string, dirName string) bool {
	isExcluded := false
	//if * match all dir contains the name
	for _, exclusiveFolder := range exclusiveFolders {
		if (strings.HasPrefix(exclusiveFolder, "*") && exclusiveFolder[1:] == dirName) ||
			exclusiveFolder == relativePath[1:] {
			isExcluded = true
		}
	}

	return isExcluded
}

func main() {
	err := loadConfigs()
	if err != nil {
		fmt.Println("Error loading configs:", err)
		os.Exit(1)
	}

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	r := mux.NewRouter()
	r.HandleFunc("/", projectHandler)
	r.HandleFunc("/p/{project_json_name}", projectHandler)
	r.HandleFunc("/v/{project_json_name}/{relativePath:.*}", fileViewHandler)
	r.HandleFunc("/f/{project_json_name}/{relativePath:.*}", fileHandler)
	r.HandleFunc("/j/{project_json_name}/{relativePath:.*}", jsonFileHandler)
	r.HandleFunc("/s/{project_json_name}/{relativePath:.*}", dirStructureHandler)
	r.HandleFunc("/c/{project_json_name}/{relativePath:.*}", dirContentsHandler)
	http.Handle("/", r)

	fmt.Println("Server is running on http://localhost:" + generalSettings.ServerPort)
	http.ListenAndServe(":"+generalSettings.ServerPort, nil)
}
