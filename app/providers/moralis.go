package providers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IPFSFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type IPFSUploader struct {
	APIKey string
	Client *http.Client
	APIUrl string
}

type IPFSFilesResponse []struct {
	Path string `json:"path"`
}

var Uploader = IPFSUploader{
	Client: &http.Client{},
	APIUrl: "https://deep-index.moralis.io/api/v2/ipfs/uploadFolder",
}

func UploadSingleFile(c *gin.Context) {
	if Uploader.APIKey == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "MORALIS_API_KEY must be set",
		})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to parse file: %s", err),
		})
		return
	}

	ipfsFile, err := prepareFileForUpload(file)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to prepare file for upload: %s", err),
		})
		return
	}

	filesResp, err := uploadToIPFS([]*IPFSFile{ipfsFile})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s", err)})
		return
	}

	pathToUploadedFile := filesResp[0].Path
	c.IndentedJSON(http.StatusCreated, gin.H{
		"message": "IPFS upload successful",
		"path":    pathToUploadedFile,
	})
}

func UploadMultipleFiles(c *gin.Context) {
	if Uploader.APIKey == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "MORALIS_API_KEY must be set",
		})
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to parse form: %s", err),
		})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	ipfsFiles, err := prepareMultipleFilesForUpload(files)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to prepare files for upload: %s", err),
		})
		return
	}

	filesResp, err := uploadToIPFS(ipfsFiles)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s", err)})
		return
	}

	pathsToUploadedFiles := make([]string, 0)
	for _, fileResp := range filesResp {
		pathsToUploadedFiles = append(pathsToUploadedFiles, fileResp.Path)
	}
	c.IndentedJSON(http.StatusCreated, gin.H{
		"message": "IPFS upload successful",
		"paths":   pathsToUploadedFiles,
	})
}

func uploadToIPFS(ipfsFiles []*IPFSFile) (IPFSFilesResponse, error) {
	requestBody, err := json.Marshal(ipfsFiles)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to marshal request body: %s", err))
	}

	req, err := http.NewRequest("POST", Uploader.APIUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to create HTTP request: %s", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", Uploader.APIKey)

	resp, err := Uploader.Client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to make HTTP request: %s", err))
	}

	defer resp.Body.Close()

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read response body: %s", err))
	}

	log.Printf("response data: %s, status code: %d", string(responseBytes), resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(fmt.Sprintf("API call failed with status code %d", resp.StatusCode))
	}

	var filesResp IPFSFilesResponse
	if err := json.Unmarshal(responseBytes, &filesResp); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to parse IPFS response: %s", err))
	}
	if len(filesResp) == 0 {
		return nil, errors.New(fmt.Sprintf("No files were uploaded: %s", err))
	}
	return filesResp, nil
}

func prepareFileForUpload(file *multipart.FileHeader) (*IPFSFile, error) {
	openedFile, err := file.Open()
	if err != nil {
		return nil, errors.New("failed to open file")
	}
	defer openedFile.Close()

	fileBytes, err := ioutil.ReadAll(openedFile)
	if err != nil {
		return nil, errors.New("failed to read file")
	}

	fileBase64 := base64.StdEncoding.EncodeToString(fileBytes)

	ipfsFile := &IPFSFile{
		Path:    file.Filename,
		Content: fileBase64,
	}

	return ipfsFile, nil
}

func prepareMultipleFilesForUpload(files []*multipart.FileHeader) ([]*IPFSFile, error) {
	var ipfsFiles []*IPFSFile
	for _, file := range files {
		ipfsFile, err := prepareFileForUpload(file)
		if err != nil {
			fmt.Println(err)
			continue
		}
		ipfsFiles = append(ipfsFiles, ipfsFile)
	}
	if len(ipfsFiles) == 0 {
		return nil, errors.New("no valid files to upload")
	}

	return ipfsFiles, nil
}
