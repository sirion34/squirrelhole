package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	files = make(map[string]string) // filename -> password
	mu    sync.Mutex
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseMultipartForm(10 << 20) // 10 MB limit

		password := r.FormValue("password")
		file, fileHeader, err := r.FormFile("file")
		extension := r.FormValue("extension")
		text := r.FormValue("text")

		var filename string
		var data []byte

		if file != nil && text != "" {
			http.Error(w, "Ошибка: вы попытались загрузить и файл и текст 💀", http.StatusBadRequest)
			return
		} else if text != "" {
			filename = fmt.Sprintf("uploads/%s.%s", text, extension)
			data = []byte(text)
		} else if file != nil {
			filename = fmt.Sprintf("uploads/%s", fileHeader.Filename)
			data, err = ioutil.ReadAll(file)

			if err != nil {
				http.Error(w, "Ошибка при чтении файла", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Ошибка: ничего не загружено 💀", http.StatusBadRequest)
			return
		}

		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			http.Error(w, "Ошибка при сохранении файла", http.StatusInternalServerError)
			return
		}

		go func(filePath string) {
			time.Sleep(5 * time.Minute)
			os.Remove(filePath)
			fmt.Printf("Файл %s удален\n", filePath)
		}(filename)

		mu.Lock()
		files[password] = filename
		mu.Unlock()

		fmt.Fprintf(w, "Файл успешно загружен с паролем: %s", password)
	} else {
		http.ServeFile(w, r, "upload.html")
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		password := r.FormValue("password")

		mu.Lock()
		file, exists := files[password]
		mu.Unlock()

		if !exists {
			http.Error(w, "Неверный пароль или файл не найден", http.StatusForbidden)
			return
		}

		http.ServeFile(w, r, file)
	} else {
		http.ServeFile(w, r, "download.html")
	}
}

func main() {
	os.MkdirAll("uploads", os.ModePerm)

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download", downloadHandler)

	fmt.Println("Сервер запущен на: 1872")
	http.ListenAndServe(":1872", nil)
}
