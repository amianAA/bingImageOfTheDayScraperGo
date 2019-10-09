package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"
)

const PATH = "Fotos/Wallpapers/" // Path inside $HOME folder to store downloaded wallpapers
const AMOUNT = 10                // Amount of pictures --> elapsed days
var MARKETS = [58]string{"ar-XA", "bg-BG", "cs-CZ", "da-DK", "de-AT", "de-CH", "de-DE", "el-GR", "en-AU", "en-CA",
	"en-GB", "en-ID", "en-IE", "en-IN", "en-MY", "en-NZ", "en-PH", "en-SG", "en-US", "en-XA", "en-ZA", "es-AR", "es-CL",
	"es-ES", "es-MX", "es-US", "es-XL", "et-EE", "fi-FI", "fr-BE", "fr-CA", "fr-CH", "fr-FR", "he-IL", "hr-HR", "hu-HU",
	"it-IT", "ja-JP", "ko-KR", "lt-LT", "lv-LV", "nb-NO", "nl-BE", "nl-NL", "pl-PL", "pt-BR", "pt-PT", "ro-RO", "ru-RU",
	"sk-SK", "sl-SL", "sv-SE", "th-TH", "tr-TR", "uk-UA", "zh-CN", "zh-HK", "zh-TW"}
var processedHashes = map[string]bool{}

func __manageError__(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func hashFileMd5(file []byte) (string, error) {
	var returnMD5String string
	md5hash := md5.New()
	hashInBytes := md5hash.Sum(file)
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}

func hashFilePathMd5(filePath string) (string, error) {
	var returnMD5String string
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return returnMD5String, err
	}
	returnMD5String, err = hashFileMd5(file)
	__manageError__(err)
	return returnMD5String, err
}

func processMD5ForExistingFiles(folderPath string) {
	files, err := ioutil.ReadDir(folderPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(folderPath, os.ModePerm)
		files, err = ioutil.ReadDir(folderPath)
	}
	__manageError__(err)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), "jpg") {
			filePath := fmt.Sprintf("%s/%s", folderPath, file.Name())
			md5Hash, err := hashFilePathMd5(filePath)
			__manageError__(err)
			if !processedHashes[md5Hash] {
				processedHashes[md5Hash] = true
			}
		}
	}
}

func performHTTPRequest(url string) *http.Response {
	headers := map[string]string{
		"User-Agent":                "Mozilla/5.0 (X11; Linux x86_64; rv:66.0) Gecko/20100101 Firefox/66.0",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.5",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
	}
	req, err := http.NewRequest("GET", url, nil)
	__manageError__(err)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	__manageError__(err)
	return resp
}

func main() {
	url := "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=%d&mkt=%s"
	usr, err := user.Current()
	downloadCounter := 0
	__manageError__(err)
	home := usr.HomeDir
	var images Images
	folderPath := fmt.Sprintf("%s/%s", home, PATH)
	fmt.Println("Processing MD5 for all existing files... (Avoiding duplicated images)")
	processMD5ForExistingFiles(folderPath)
	for _, mkt := range MARKETS {
		completeURL := fmt.Sprintf(url, AMOUNT, mkt)
		response := performHTTPRequest(completeURL)
		body, err := ioutil.ReadAll(response.Body)
		__manageError__(err)
		err = json.Unmarshal(body, &images)
		__manageError__(err)
		for index := range images.Images {
			pic := images.Images[index]
			var title string
			completeUrl := fmt.Sprintf("https://www.bing.com%s", pic.Url)
			image := performHTTPRequest(completeUrl)
			bodyInBytes, _ := ioutil.ReadAll(image.Body)
			md5hash, err := hashFileMd5(bodyInBytes)
			__manageError__(err)
			if !processedHashes[md5hash] {
				processedHashes[md5hash] = true
				if pic.Title != "" {
					title = pic.Title
				} else {
					title = strings.ReplaceAll(pic.Copyright, "/", ", ")
					title = strings.ReplaceAll(title, "@ ", "By ")
					title = strings.ReplaceAll(title, ", Getty Images", "")
					title = strings.ReplaceAll(title, "© ", "")
					title = strings.ReplaceAll(title, "©", "")
				}
				fileName := fmt.Sprintf("%s - %s", pic.StartDate, title)
				filePath := fmt.Sprintf("%s%s.jpg", folderPath, fileName)
				_, _ = os.Stat(folderPath)
				err = ioutil.WriteFile(filePath, bodyInBytes, os.ModePerm)
				fmt.Println("Downloading... ", filePath)
				downloadCounter += 1
			}
		}
	}
	fmt.Printf("Downloaded %d new images", downloadCounter)
}
