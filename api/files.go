package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"easilypanel5/utils"
)

// FileInfo 文件信息结构
type FileInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	IsDir        bool      `json:"is_dir"`
	ModTime      time.Time `json:"mod_time"`
	Permissions  string    `json:"permissions"`
	Extension    string    `json:"extension"`
	MimeType     string    `json:"mime_type"`
	CanRead      bool      `json:"can_read"`
	CanWrite     bool      `json:"can_write"`
	CanExecute   bool      `json:"can_execute"`
}

// FileListRequest 文件列表请求
type FileListRequest struct {
	Path   string `json:"path"`
	SortBy string `json:"sort_by"` // name, size, modified
	Order  string `json:"order"`   // asc, desc
}

// FileOperationRequest 文件操作请求
type FileOperationRequest struct {
	Action string `json:"action"` // create, delete, rename, copy, move
	Path   string `json:"path"`
	Target string `json:"target,omitempty"`
	IsDir  bool   `json:"is_dir,omitempty"`
}

// handleFiles 处理文件管理请求
func handleFiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleFileList(w, r)
	case http.MethodPost:
		handleFileOperation(w, r)
	case http.MethodPut:
		handleFileUpload(w, r)
	default:
		WriteStandardError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleFileList 处理文件列表请求
func handleFileList(w http.ResponseWriter, r *http.Request) {
	// 获取路径参数
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}

	// 验证路径安全性
	if err := utils.ValidateFilePath("path", path); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			WriteStandardError(w, "VALIDATION_FAILED", valErr.Message, http.StatusBadRequest)
			return
		}
	}

	// 获取排序参数
	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "name"
	}
	order := r.URL.Query().Get("order")
	if order == "" {
		order = "asc"
	}

	// 读取目录内容
	files, err := listFiles(path, sortBy, order)
	if err != nil {
		WriteStandardError(w, "FILE_LIST_FAILED", fmt.Sprintf("Failed to list files: %v", err), http.StatusInternalServerError)
		return
	}

	WriteStandardResponse(w, map[string]interface{}{
		"path":  path,
		"files": files,
		"count": len(files),
	})
}

// handleFileOperation 处理文件操作请求
func handleFileOperation(w http.ResponseWriter, r *http.Request) {
	var req FileOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteStandardError(w, "INVALID_JSON", "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 验证路径
	if err := utils.ValidateFilePath("path", req.Path); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			WriteStandardError(w, "VALIDATION_FAILED", valErr.Message, http.StatusBadRequest)
			return
		}
	}

	// 执行操作
	switch req.Action {
	case "create":
		err := createFileOrDir(req.Path, req.IsDir)
		if err != nil {
			WriteStandardError(w, "CREATE_FAILED", fmt.Sprintf("Failed to create: %v", err), http.StatusInternalServerError)
			return
		}
	case "delete":
		err := deleteFileOrDir(req.Path)
		if err != nil {
			WriteStandardError(w, "DELETE_FAILED", fmt.Sprintf("Failed to delete: %v", err), http.StatusInternalServerError)
			return
		}
	case "rename":
		if req.Target == "" {
			WriteStandardError(w, "MISSING_TARGET", "Target path is required for rename operation", http.StatusBadRequest)
			return
		}
		err := renameFileOrDir(req.Path, req.Target)
		if err != nil {
			WriteStandardError(w, "RENAME_FAILED", fmt.Sprintf("Failed to rename: %v", err), http.StatusInternalServerError)
			return
		}
	case "copy":
		if req.Target == "" {
			WriteStandardError(w, "MISSING_TARGET", "Target path is required for copy operation", http.StatusBadRequest)
			return
		}
		err := copyFileOrDir(req.Path, req.Target)
		if err != nil {
			WriteStandardError(w, "COPY_FAILED", fmt.Sprintf("Failed to copy: %v", err), http.StatusInternalServerError)
			return
		}
	default:
		WriteStandardError(w, "INVALID_ACTION", "Invalid action", http.StatusBadRequest)
		return
	}

	WriteStandardResponse(w, map[string]string{"status": "success"})
}

// handleFileUpload 处理文件上传
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	// 解析multipart表单
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		WriteStandardError(w, "PARSE_FORM_FAILED", "Failed to parse form", http.StatusBadRequest)
		return
	}

	// 获取目标路径
	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = "."
	}

	// 验证路径
	if err := utils.ValidateFilePath("path", targetPath); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			WriteStandardError(w, "VALIDATION_FAILED", valErr.Message, http.StatusBadRequest)
			return
		}
	}

	// 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		WriteStandardError(w, "FILE_UPLOAD_FAILED", "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 创建目标文件
	targetFile := filepath.Join(targetPath, header.Filename)
	dst, err := os.Create(targetFile)
	if err != nil {
		WriteStandardError(w, "CREATE_FILE_FAILED", fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 复制文件内容
	_, err = io.Copy(dst, file)
	if err != nil {
		WriteStandardError(w, "COPY_FILE_FAILED", fmt.Sprintf("Failed to copy file: %v", err), http.StatusInternalServerError)
		return
	}

	WriteStandardResponse(w, map[string]interface{}{
		"filename": header.Filename,
		"size":     header.Size,
		"path":     targetFile,
	})
}

// listFiles 列出文件
func listFiles(path, sortBy, order string) ([]FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		fileInfo := FileInfo{
			Name:        entry.Name(),
			Path:        fullPath,
			Size:        info.Size(),
			IsDir:       entry.IsDir(),
			ModTime:     info.ModTime(),
			Permissions: info.Mode().String(),
			Extension:   filepath.Ext(entry.Name()),
			CanRead:     true,  // 简化处理
			CanWrite:    true,  // 简化处理
			CanExecute:  false, // 简化处理
		}

		// 设置MIME类型
		if !entry.IsDir() {
			fileInfo.MimeType = getMimeType(fileInfo.Extension)
		}

		files = append(files, fileInfo)
	}

	// 排序
	sortFiles(files, sortBy, order)

	return files, nil
}

// sortFiles 排序文件列表
func sortFiles(files []FileInfo, sortBy, order string) {
	sort.Slice(files, func(i, j int) bool {
		// 目录总是排在前面
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}

		var less bool
		switch sortBy {
		case "size":
			less = files[i].Size < files[j].Size
		case "modified":
			less = files[i].ModTime.Before(files[j].ModTime)
		default: // name
			less = strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
		}

		if order == "desc" {
			return !less
		}
		return less
	})
}

// getMimeType 获取MIME类型
func getMimeType(ext string) string {
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".md":   "text/markdown",
		".json": "application/json",
		".xml":  "application/xml",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".jar":  "application/java-archive",
	}

	if mimeType, exists := mimeTypes[strings.ToLower(ext)]; exists {
		return mimeType
	}
	return "application/octet-stream"
}

// createFileOrDir 创建文件或目录
func createFileOrDir(path string, isDir bool) error {
	if isDir {
		return os.MkdirAll(path, 0755)
	} else {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		return file.Close()
	}
}

// deleteFileOrDir 删除文件或目录
func deleteFileOrDir(path string) error {
	return os.RemoveAll(path)
}

// renameFileOrDir 重命名文件或目录
func renameFileOrDir(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// copyFileOrDir 复制文件或目录
func copyFileOrDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(src, dst)
	} else {
		return copyFile(src, dst)
	}
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyDir 复制目录
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			return copyFile(path, dstPath)
		}
	})
}

// handleFileDownload 处理文件下载
func handleFileDownload(w http.ResponseWriter, r *http.Request) {
	if err := ValidateMethod(r, http.MethodGet); err != nil {
		WriteStandardError(w, "METHOD_NOT_ALLOWED", err.Error(), http.StatusMethodNotAllowed)
		return
	}

	// 获取文件路径
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		WriteStandardError(w, "MISSING_PATH", "File path is required", http.StatusBadRequest)
		return
	}

	// 验证路径
	if err := utils.ValidateFilePath("path", filePath); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			WriteStandardError(w, "VALIDATION_FAILED", valErr.Message, http.StatusBadRequest)
			return
		}
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		WriteStandardError(w, "FILE_NOT_FOUND", "File not found", http.StatusNotFound)
		return
	}

	if fileInfo.IsDir() {
		WriteStandardError(w, "IS_DIRECTORY", "Cannot download directory", http.StatusBadRequest)
		return
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		WriteStandardError(w, "OPEN_FILE_FAILED", fmt.Sprintf("Failed to open file: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 设置响应头
	filename := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// 发送文件内容
	io.Copy(w, file)
}

// handleFileContent 处理文件内容查看/编辑
func handleFileContent(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetFileContent(w, r)
	case http.MethodPost:
		handleSaveFileContent(w, r)
	default:
		WriteStandardError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetFileContent 获取文件内容
func handleGetFileContent(w http.ResponseWriter, r *http.Request) {
	// 获取文件路径
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		WriteStandardError(w, "MISSING_PATH", "File path is required", http.StatusBadRequest)
		return
	}

	// 验证路径
	if err := utils.ValidateFilePath("path", filePath); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			WriteStandardError(w, "VALIDATION_FAILED", valErr.Message, http.StatusBadRequest)
			return
		}
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		WriteStandardError(w, "FILE_NOT_FOUND", "File not found", http.StatusNotFound)
		return
	}

	if fileInfo.IsDir() {
		WriteStandardError(w, "IS_DIRECTORY", "Cannot read directory content", http.StatusBadRequest)
		return
	}

	// 检查文件大小（限制为1MB）
	if fileInfo.Size() > 1024*1024 {
		WriteStandardError(w, "FILE_TOO_LARGE", "File too large to display", http.StatusBadRequest)
		return
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		WriteStandardError(w, "READ_FILE_FAILED", fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		return
	}

	WriteStandardResponse(w, map[string]interface{}{
		"path":     filePath,
		"content":  string(content),
		"size":     fileInfo.Size(),
		"modified": fileInfo.ModTime(),
	})
}

// SaveFileContentRequest 保存文件内容请求
type SaveFileContentRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// handleSaveFileContent 保存文件内容
func handleSaveFileContent(w http.ResponseWriter, r *http.Request) {
	var req SaveFileContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteStandardError(w, "INVALID_JSON", "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 验证路径
	if err := utils.ValidateFilePath("path", req.Path); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			WriteStandardError(w, "VALIDATION_FAILED", valErr.Message, http.StatusBadRequest)
			return
		}
	}

	// 写入文件内容
	err := os.WriteFile(req.Path, []byte(req.Content), 0644)
	if err != nil {
		WriteStandardError(w, "WRITE_FILE_FAILED", fmt.Sprintf("Failed to write file: %v", err), http.StatusInternalServerError)
		return
	}

	WriteStandardResponse(w, map[string]string{"status": "success"})
}
