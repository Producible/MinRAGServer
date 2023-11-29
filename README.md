
# MinFileServer

The MinFileServer is a Go-based web application that provides a user-friendly interface to navigate through project directories and view file contents. It's designed to assist ChatGPT in reading files from URLs, with the support of scraping plugins. The application allows users quickly to browse through different projects, expand/collapse directories, and view file contents in a new window.

## Features
- Display a list of projects configured in JSON files.
- Navigate through directories and subdirectories of a project.
- View file contents in a new window.
- Copy file URLs and info to the clipboard to paste in ChatGPT
- Configure visibility settings such as showing/hiding hidden files, timestamping URLs, and filtering files and folders based on extensions or names.
- Customize appearance through CSS and JavaScript.
- Supports scraping plugins for enhanced functionality.

## Installation
1. Ensure you have Go installed on your machine. You can download it from the official website.

2. Clone this repository to your local machine.
```
git clone https://github.com/greatwhiz/MinFileServer.git
cd MinFileServer
```

3. Build the project.
```
go build
```

For Windows:
```
GOOS=windows GOARCH=amd64 go build
```

For Mac:
```
GOOS=darwin GOARCH=amd64 go build
```

## Configuration
1. Configure the general settings in settings.json:
```
{
  "server_port": "8080",
  "disable_external_network_browsing": true,
  "show_hidden": false,
  "time_stamp": true,
  "inclusive_extensions": "*",
  "exclusive_extensions": "",
  "exclusive_folders": ""  
}
```
2. Create a folder named config and inside it, create a JSON file for each project you want to display. The JSON file should have the following structure:
```
{
    "project_name": "Project 1",
    "root_path": "/path/to/project",
    "project_url": "http://external-domain:80",
    "inclusive_extensions": "js,ts,tsx,json,css,html",
    "exclusive_extensions": "",
    "exclusive_folders": "node_modules,build,dist,coverage"  
}
```
The project_url includes the host and the port which can be access from Internet. You can use dynamic dns and port mapping to your local network.

## Usage
1. Run the server:
```
./furl
```
2. Open your web browser and navigate to http://localhost:8080.
3. Click on a project name to view its file tree.
4. Navigate through the directories by clicking on the folder icons.
5. Click on a file name to view its content in a new window.
6. Use the copy icons to copy the file URL or info to the clipboard.
7. With any scraper plugin you prefer, paste the URL in the prompt.

## Customization
- Customize the appearance by modifying the static/style.css file.
- Add interactive features by modifying the static/script.js file.
- The static/clipboard.js file handles the copy-to-clipboard functionality.

## Contributing
Feel free to fork this repository, make changes, and open a pull request. Contributions are welcome!

## License
This project is licensed under the MIT License. See the LICENSE file for details.
