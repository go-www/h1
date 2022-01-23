package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
)

type ImgBBAPIResponse struct {
	Data    Data `json:"data"`
	Success bool `json:"success"`
	Status  int  `json:"status"`
}
type Image struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}
type Thumb struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}
type Medium struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}
type Data struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	URLViewer  string `json:"url_viewer"`
	URL        string `json:"url"`
	DisplayURL string `json:"display_url"`
	Size       string `json:"size"`
	Time       string `json:"time"`
	Expiration string `json:"expiration"`
	Image      Image  `json:"image"`
	Thumb      Thumb  `json:"thumb"`
	Medium     Medium `json:"medium"`
	DeleteURL  string `json:"delete_url"`
}

func uploadImage(imageData []byte, imageName string) (imgURL string, err error) {
	var buffer bytes.Buffer
	multipartWriter := multipart.NewWriter(&buffer)
	multipartWriter.WriteField("key", os.Getenv("IMGBB_API_KEY"))
	imgFileWriter, err := multipartWriter.CreateFormFile("image", imageName)
	if err != nil {
		return
	}
	_, err = imgFileWriter.Write(imageData)
	if err != nil {
		return
	}
	multipartWriter.Close()
	resp, err := http.Post("https://api.imgbb.com/1/upload", multipartWriter.FormDataContentType(), &buffer)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var imgBBAPIResponse ImgBBAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&imgBBAPIResponse)
	if err != nil {
		return
	}
	if !imgBBAPIResponse.Success {
		err = fmt.Errorf("imgBB API error: %v", imgBBAPIResponse.Status)
	}
	return imgBBAPIResponse.Data.URL, nil
}
