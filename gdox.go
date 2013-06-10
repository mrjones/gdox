package main

import (
	oauth "code.google.com/p/goauth2/oauth"
	drive "code.google.com/p/google-api-go-client/drive/v2"

	"../geyefi"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Config struct {
	GoogleClientId string
	GoogleClientSecret string
	EyeFiUploadKey string
}

type GDocsHandler struct {
	service *drive.Service
}

func NewGDocsHandler(config *Config) (*GDocsHandler, error) {
	httpClient, err := authorize(config);
	if err != nil {
		return nil, err
	}


	service, err := drive.New(httpClient)
	if err != nil {
		return nil, err
	}

	return &GDocsHandler{service: service}, nil
}

func (g *GDocsHandler) HandleUpload(filename string, data []byte) error {
	driveFile := &drive.File{Title: filename}
	parent := &drive.ParentReference{Id: "0B1SxUBEP5_X2ZEdMaW45Qy1KcFk"}
	driveFile.Parents = []*drive.ParentReference{parent}

	_, err := g.service.Files.Insert(driveFile).Ocr(true).OcrLanguage("en").Media(bytes.NewReader(data)).Do();
	if err != nil {
		return err
	}

	return nil	
}

func parseConfigFile(filename string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var configFile Config;
	err = json.Unmarshal(bytes, &configFile);
	if err != nil {
		return nil, err
	}

	return &configFile, nil
}

func authorize(configFile *Config) (*http.Client, error) {
	config := &oauth.Config{
		ClientId: configFile.GoogleClientId,
		ClientSecret: configFile.GoogleClientSecret,
		Scope: "https://www.googleapis.com/auth/drive",
		AuthURL: "https://accounts.google.com/o/oauth2/auth",
		TokenURL: "https://accounts.google.com/o/oauth2/token",
		RedirectURL: "oob",
	}

	transport := &oauth.Transport{Config: config}

	tokenCache := oauth.CacheFile("tokens.cache")

	token, err := tokenCache.Token()
	if err != nil {
		url := config.AuthCodeURL("")
		fmt.Println("Visit this URL to get a code, then type it in below:")
		fmt.Println(url)

		verificationCode := ""
		fmt.Scanln(&verificationCode)

		token, err := transport.Exchange(verificationCode)
		if err != nil {
			return nil, err
		}

		err = tokenCache.PutToken(token)
		if err != nil {
			log.Printf("Error: %s\n", err)
		}
	} else {
		transport.Token = token
	}

	return transport.Client(), nil
}

func main() {
	config, err := parseConfigFile("gdox.conf.json")
	if err != nil {
		log.Fatal(err)
	}

	handler, err := NewGDocsHandler(config)
	if err != nil {
		log.Fatal(err)
	}

	e := geyefi.NewServer("818b6183a1a0839d88366f5d7a4b0161", handler)
	e.ListenAndServe()
}
