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
	GoogleAPIKey string
	GCMRegistrationId string
}

type GDocsHandler struct {
	service *drive.Service
	googleApiKey string
	gcmRegistrationId string
}

type BeepBoopData struct {
	Title string `json:"title"`
	Body string `json:"body"`
}

type GcmRequest struct {
	RegistrationIds []string `json:"registration_ids"`
	Data BeepBoopData `json:"data,omitempty"`
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

	return &GDocsHandler{service: service, googleApiKey: config.GoogleAPIKey, gcmRegistrationId: config.GCMRegistrationId}, nil
}

func (g *GDocsHandler) HandleUpload(filename string, data []byte) error {
	driveFile := &drive.File{Title: filename}
	// TODO: make configurable
	parent := &drive.ParentReference{Id: "0B1SxUBEP5_X2ZEdMaW45Qy1KcFk"}
	driveFile.Parents = []*drive.ParentReference{parent}

	_, err := g.service.Files.Insert(driveFile).Ocr(true).OcrLanguage("en").Media(bytes.NewReader(data)).Do();
	if err != nil {
		return err
	}

	if (g.googleApiKey != "") {
		err = g.sendNotification("GDox", "File uploaded")
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (g *GDocsHandler) sendNotification(title, body string) error {
	// TODO: handle on "last photo in roll"?
	gcmRequest := GcmRequest{}
	gcmRequest.RegistrationIds = make([]string, 1)
	gcmRequest.RegistrationIds[0] = g.gcmRegistrationId
	gcmRequest.Data.Title = title
	gcmRequest.Data.Body = body

	reqBody, err := json.Marshal(gcmRequest)
	if err != nil {
		return err
	}
	log.Println("Sending: " + string(reqBody))

	req, err := http.NewRequest(
		"POST", "https://android.googleapis.com/gcm/send", bytes.NewReader(reqBody))

	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "key=" + g.googleApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
		
	if err != nil {
		return err
	}

	log.Println(string(respBody))
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

	e := geyefi.NewServer(config.EyeFiUploadKey, handler)
	e.ListenAndServe()
}
