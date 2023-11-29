package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
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

	fmt.Fprintln(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>MinFileServer</title>
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
	vars := mux.Vars(r)
	project := vars["project_json_name"]
	path := vars["relativePath"]

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
	vars := mux.Vars(r)
	project := vars["project_json_name"]
	path := vars["relativePath"]

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

			fileLink := fmt.Sprintf("f/%s/%s", project, relativePath)
			url := fmt.Sprintf("%s/%s", selectedConfig.ProjectURL, fileLink)
			info := fmt.Sprintf("%s: %s", file.Name(), url)
			jsonLink := fmt.Sprintf("j/%s/%s", project, relativePath)
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
	r := mux.NewRouter()
	r.HandleFunc("/", projectHandler)
	r.HandleFunc("/p/{project_json_name}", projectHandler)
	r.HandleFunc("/f/{project_json_name}/{relativePath:.*}", fileHandler)
	r.HandleFunc("/j/{project_json_name}/{relativePath:.*}", jsonFileHandler)
	http.Handle("/", r)

	fmt.Println("Server is running on http://localhost:" + generalSettings.ServerPort)
	http.ListenAndServe(":"+generalSettings.ServerPort, nil)
}
