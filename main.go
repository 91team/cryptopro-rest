package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	apiKey := getEnv("API_KEY", "")
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	http.HandleFunc("/api/sign", func(w http.ResponseWriter, r *http.Request) {
		if "Bearer "+apiKey != r.Header.Get("Authorization") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		r.Header.Set("Content-Type", "text/plain")
		w.Write(signData(body))
	})
	port := getEnv("PORT", "3000")
	log.Printf("Signer server started on port: %v", port)
	http.ListenAndServe(":"+port, nil)
}

func signData(data []byte) []byte {
	dir := "./tmp/"
	filename := genFileName()
	path := dir + filename
	ioutil.WriteFile(path, data, 0644)

	cmd := exec.Command(
		"cryptcp", "-signf", "-cert", "-detached", "-nochain",
		"-thumbprint", os.Getenv("KEY_THUMBPRINT"),
		"-pin", os.Getenv("KEY_PASSWORD"),
		"-dir", dir,
		path,
	)
	_, err := cmd.Output()

	if err != nil {
		log.Println(err.Error())
		return []byte{}
	}

	content, _ := ioutil.ReadFile(path + ".sgn")
	os.Remove(path)
	os.Remove(path + ".sgn")

	return []byte(strings.TrimSuffix(string(content), "\n"))
}

func genFileName() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 12
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
