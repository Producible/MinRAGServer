package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type GeneralSettings struct {
	ServerPort          string `json:"server_port"`
	ShowHidden          bool   `json:"show_hidden"`
	TimeStamp           bool   `json:"time_stamp"`
	InclusiveExtensions string `json:"inclusive_extensions"`
	ExclusiveExtensions string `json:"exclusive_extensions"`
	ExclusiveFolders    string `json:"exclusive_folders"`
}

var generalSettings GeneralSettings

type Config struct {
	ProjectName         string `json:"project_name"`
	RootPath            string `json:"root_path"`
	ProjectURL          string `json:"project_url"`
	InclusiveExtensions string `json:"inclusive_extensions,omitempty"`
	ExclusiveExtensions string `json:"exclusive_extensions,omitempty"`
	ExclusiveFolders    string `json:"exclusive_folders,omitempty"`
}

var configs map[string]Config
var selectedConfig Config

func loadConfigs() error {
	// Load general settings
	settingsData, err := ioutil.ReadFile("settings.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(settingsData, &generalSettings)
	if err != nil {
		return err
	}

	configs = make(map[string]Config)
	files, err := ioutil.ReadDir("config")
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
	project := r.URL.Query().Get("p")
	if project == "" || configs[project+".json"].ProjectName == "" {
		for filename, config := range configs {
			projectID := strings.TrimSuffix(filename, ".json")
			fmt.Fprintf(w, "<a href='/?p=%s'>%s</a><br>", projectID, config.ProjectName)
		}
		return
	}

	selectedConfig = configs[project+".json"]
	dirPath := selectedConfig.RootPath

	fmt.Fprintln(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>FileToURLs</title>
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
<li class="root-item expanded"><div class='item'><span>`+selectedConfig.ProjectName+`</span></div><ul>`)

	writeDirectory(w, dirPath, dirPath, project)

	fmt.Fprintln(w, `</ul></li>  <!-- Updated line -->
</ul>
</div>
<script src="/static/script.js"></script>
<script src="/static/clipboard.js"></script>
</body>
</html>`)
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("p")
	path := r.URL.Query().Get("path")

	if project == "" || configs[project+".json"].ProjectName == "" {
		http.Error(w, "Invalid project", http.StatusBadRequest)
		return
	}

	selectedConfig = configs[project+".json"]
	fullPath := filepath.Join(selectedConfig.RootPath, path)

	data, err := ioutil.ReadFile(fullPath)
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

func jsonFileHandler(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("p")
	path := r.URL.Query().Get("path")

	if project == "" || configs[project+".json"].ProjectName == "" {
		http.Error(w, "Invalid project", http.StatusBadRequest)
		return
	}

	selectedConfig = configs[project+".json"]
	fullPath := filepath.Join(selectedConfig.RootPath, path)

	data, err := ioutil.ReadFile(fullPath)
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
		"file": fileName,
		"path": relativePath,
		"data": jsonData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func writeDirectory(w http.ResponseWriter, path string, rootPath string, project string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Separate directories and files
	var dirs []os.FileInfo
	var filesOnly []os.FileInfo
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

	// Inside the writeDirectory function
	for _, dir := range dirs {
		// Skip the folder if it's in the exclusive_folders list
		if contains(exclusiveFolders, dir.Name()) {
			continue
		}

		fmt.Fprintf(w, "<li><div class='item'><span>%s</span></div><ul>", dir.Name())
		writeDirectory(w, filepath.Join(path, dir.Name()), rootPath, project) // Recursive call
		fmt.Fprintln(w, "</ul></li>")
	}

	// Process files
	for _, file := range filesOnly {

		ext := filepath.Ext(file.Name()) // Get file extension without the leading "."

		// when ext not empty, remove "."
		if len(ext) > 0 {
			ext = ext[1:]
		}

		if (len(inclusiveExtensions) == 0 || inclusiveExtensions[0] == "*" || contains(inclusiveExtensions, ext)) &&
			(len(exclusiveExtensions) == 0 || !contains(exclusiveExtensions, ext)) {
			relativePath := strings.TrimPrefix(filepath.Join(path, file.Name()), rootPath)
			relativePath = filepath.ToSlash(relativePath)
			url := fmt.Sprintf("%s/%s", selectedConfig.ProjectURL, strings.TrimPrefix(relativePath, "/"))
			info := fmt.Sprintf("%s: %s", file.Name(), url)

			fileLink := fmt.Sprintf("/file?p=%s&path=%s", project, relativePath)
			jsonLink := fmt.Sprintf("/jsonfile?p=%s&path=%s", project, relativePath)
			fmt.Fprintf(w, "<li><div class='item'><a href='%s' target='_blank'>%s</a> <a href='%s' target='_blank' class='buttons'><i class='fas fa-external-link-alt'></i></a> <button class='copy-button buttons' data-url='%s'><i class='fas fa-copy'></i></button> <button class='copy-button-info buttons' data-info='%s'><i class='fas fa-copy'></i></button> <a href='%s' target='_blank' class='buttons'><i class='fas fa-file-code'></i></a></div></li>", fileLink, file.Name(), url, url, info, jsonLink)
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

func main() {
	err := loadConfigs()
	if err != nil {
		fmt.Println("Error loading configs:", err)
		os.Exit(1)
	}

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", projectHandler)
	http.HandleFunc("/file", fileHandler)
	http.HandleFunc("/jsonfile", jsonFileHandler)

	fmt.Println("Server is running on http://localhost:" + generalSettings.ServerPort)
	http.ListenAndServe(":"+generalSettings.ServerPort, nil)
}
