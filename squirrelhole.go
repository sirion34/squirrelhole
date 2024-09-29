package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	files = make(map[string]string)
	mu    sync.Mutex
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomString(length int) string {
	// For random naming file with text
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func hashNameGenerate(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	hashedBytes := hash.Sum(nil)
	hashedString := hex.EncodeToString(hashedBytes)
	return hashedString[:32]
}

func setFileName(password string, filename string) string {
	_, exists := files[password]
	if exists {
		data := []byte("❗❗❗password compromised❗❗❗")
		os.Remove(filename)
		encrypt(data, filename)
		return "Password " + fmt.Sprint(password) + " is used by another User"
	} else {
		files[password] = filename
		return "File successfully uploaded with password: " + fmt.Sprint(password)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseMultipartForm(10 << 20) // 10 MB limit

		password := r.FormValue("password")
		file, _, err := r.FormFile("file")
		text := r.FormValue("text")

		var filename string
		var data []byte
		filename = fmt.Sprintf("uploads/%s", hashNameGenerate(password))

		if file != nil && text != "" {
			http.Error(w, "Error: you tried to upload both file and text 💀", http.StatusBadRequest)
			return
		} else if text != "" {
			data = []byte(text)
		} else if file != nil {
			data, err = ioutil.ReadAll(file)

			if err != nil {
				http.Error(w, "Error reading file", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Error: nothing uploaded 💀", http.StatusBadRequest)
			return
		}

		encrypt(data, filename)

		mu.Lock()
		result := setFileName(password, filename)
		mu.Unlock()

		go func(password string, filePath string) {
			time.Sleep(1 * time.Minute)
			delete(files, password)
			os.Remove(filePath)
			fmt.Printf("File %s has been deleted\n", filePath)
		}(password, filename)

		fmt.Fprintf(w, result)
	} else {
		http.ServeFile(w, r, "src/upload.html")
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		password := r.FormValue("password")

		mu.Lock()
		file, exists := files[password]
		mu.Unlock()

		plaintext, err := decrypt(file)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "Incorrect password or file not found", http.StatusForbidden)
			return
		}

		delete(files, password)
		os.Remove(file)

		fmt.Fprintf(w, "%s", string(plaintext))
	} else {
		http.ServeFile(w, r, "src/download.html")
	}
}

func homepage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "src/homepage.html")
}

func main() {
	os.MkdirAll("uploads", os.ModePerm)

	http.HandleFunc("/", homepage)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download", downloadHandler)

	fmt.Println("Server start: 1872")
	http.ListenAndServe(":1872", nil)
}
