
# MinRAGServer

MinRAGServer is a Go-based web application designed to enhance the capabilities of Retrieval Augmented Generation (RAG) models like ChatGPT. With its user-friendly interface, MinRAGServer simplifies the process of navigating project directories and viewing file contents, making it an invaluable tool for feeding content to ChatGPT, especially when used with scraping plugins. Users can quickly browse through different projects, expand or collapse directories, and view file contents in new windows, streamlining the data retrieval process for RAG models.

<img width="600" alt="image_2023-11-29_13-45-03" src="https://github.com/greatwhiz/MinFileServer/assets/35230556/6e70a6f5-482b-49b3-b3cd-571e73005564">

<img width="600" alt="image_2023-11-29_16-45-03" src="https://github.com/greatwhiz/MinFileServer/assets/35230556/1c6ede8b-8833-40f6-95d5-d53c3e32726b">

<img width="600" alt="image_2023-11-29_16-47-06" src="https://github.com/greatwhiz/MinFileServer/assets/35230556/b4f21a43-5d78-4420-84ca-e1cfa0e8982c">

We also highly recommend using the Chrome extension ChatGPT Helper alongside MinRAGServer to further increase productivity.
https://chromewebstore.google.com/detail/chatgpt-helper/pjaiffleeblodclagbgflpnmighceibl?hl=en

## Features
- Display and navigate a list of projects configured in JSON files.
- View the entire structure and all file contents of a project in new windows.
- Copy structure URLs and content URLs to the clipboard for use in ChatGPT.
- Configure visibility settings, including show/hide hidden files, timestamp URLs, and filter files and folders based on extensions or names.
- Customize appearance and add interactive features through CSS and JavaScript.
- Support for scraping plugins to enhance RAG model functionality.

## Installation
1. Ensure you have Go installed on your machine. You can download it from the official website.

2. Clone this repository to your local machine.
```
git clone https://github.com/greatwhiz/MinRAGServer.git
cd MinRAGServer
```

3. Build the project.
```
go mod tidy
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
  "inclusive_extensions": "js,ts,tsx,json,css,html",
  "exclusive_extensions": "",
  "exclusive_folders": "node_modules,build,dist,coverage"  
}
```
2. Create a folder named config and inside it, create a JSON file for each project you want to display. The JSON file should have the following structure:
```
{
    "project_name": "Project 1",
    "root_path": "/absolute_path/to/project",
    "project_url": "http://external-domain:80",
    "inclusive_extensions": "js,ts,tsx,json,css,cs,html,dart",
    "exclusive_extensions": "",
    "exclusive_folders": "linux,macos,windows,build",
    "exclusive_files": ""
}
```
The project_url includes the host and the port which can be accessed from the Internet. You can use dynamic DNS and port mapping to your local network.

## Usage
1. Run the server:
```
./MinRAGServer
```
2. Open a web browser and navigate to http://localhost:8080.
3. Map an external port on your router if necessary
4. Click on a project name to view its file tree.
5. Navigate through directories and view file contents.
6. Copy file URLs and info to the clipboard.
7. Use a scraper plugin to feed content to ChatGPT.
8. (Optional) Enhance productivity with the ChatGPT Helper Chrome extension.
https://chromewebstore.google.com/detail/chatgpt-helper/pjaiffleeblodclagbgflpnmighceibl?hl=en

## Customization
- Modify static/style.css to customize the appearance.
- Add features with static/script.js.
- static/clipboard.js handles copy-to-clipboard functionality.

## Contributing
Contributions are welcome! Fork the repository, make changes, and open a pull request.

## License
MinRAGServer is licensed under the MIT License. See the LICENSE file for details.
