package main

import (
	"encoding/base64"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var CertHash = ""
var SignAlg = ""

func main() {
	apiKey := getEnv("API_KEY", "")
	calcCertHash()
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	http.HandleFunc("/api/sign", func(w http.ResponseWriter, r *http.Request) {
		if "Bearer "+apiKey != r.Header.Get("Authorization") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		r.Header.Set("Content-Type", "text/plain")
		w.Write(signData(body))
	})
	http.HandleFunc("/api/sign-esia", func(w http.ResponseWriter, r *http.Request) {
		if "Bearer "+apiKey != r.Header.Get("Authorization") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		r.Header.Set("Content-Type", "text/plain")
		w.Write([]byte(signEsia(body)))
	})
	port := getEnv("PORT", "3000")
	log.Printf("Signer server started on port: %v", port)
	http.ListenAndServe(":"+port, nil)
}

func signEsia(data []byte) string {
	dir := "./tmp/"
	filename := genFileName()
	path := dir + filename
	os.WriteFile(path, data, 0644)

	cmd := exec.Command(
		"csptest", "-keyset",
		"-container", os.Getenv("CONTAINER"),
		"-password", os.Getenv("KEY_PASSWORD"),
		"-sign", SignAlg,
		"-keytype", "exchange",
		"-in", path,
		"-out", path+".sgn",
	)
	_, err := cmd.Output()

	if err != nil {
		log.Println(err.Error())
		return ""
	}

	content, _ := os.ReadFile(path + ".sgn")
	os.Remove(path)
	os.Remove(path + ".sgn")
	return base64.URLEncoding.EncodeToString(reverse(content))
}

func signData(data []byte) []byte {
	dir := "./tmp/"
	filename := genFileName()
	path := dir + filename
	os.WriteFile(path, data, 0644)

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

	content, _ := os.ReadFile(path + ".sgn")
	os.Remove(path)
	os.Remove(path + ".sgn")

	return []byte(strings.TrimSuffix(string(content), "\n"))
}

func calcCertHash() {
	var alg string

	cmd := exec.Command(
		"certmgr", "-export",
		"-thumbprint", os.Getenv("KEY_THUMBPRINT"),
		"-dest", "cert.cer",
	)
	_, err := cmd.Output()

	if err != nil {
		log.Println(err.Error())
		return
	}

	// openssl x509 -in $2 -inform der -text | grep 'Signature Algorithm' -m 1 | xargs
	cmd = exec.Command("bash", "-c", "openssl x509 -in cert.cer -inform der -text | grep 'Signature Algorithm' -m 1 | xargs")
	algOut, err := cmd.Output()
	if err != nil {
		log.Println(err.Error())
		return
	}
	switch string(algOut) {
	case "Signature Algorithm: GOST R 34.10-2012 with GOST R 34.11-2012 (256 bit)\n":
		alg = "GR3411_2012_256"
		SignAlg = "GOST12_256"
	case "Signature Algorithm: GOST R 34.10-2012 with GOST R 34.11-2012 (512 bit)\n":
		alg = "GR3411_2012_512"
		SignAlg = "GOST12_512"
	default:
		alg = "unknown"
		return
	}

	cmd = exec.Command(
		"cpverify", "cert.cer", "-mk", "-alg", alg, "-inverted_halfbytes", "0",
	)
	hashOut, err := cmd.Output()
	if err != nil {
		log.Println(err.Error())
		return
	}
	CertHash = string(hashOut)
	log.Printf("Certificate hash(%s): %s", alg, CertHash)
}

func genFileName() string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
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

func reverse(input []byte) []byte {
	inputLen := len(input)
	output := make([]byte, inputLen)

	for i, n := range input {
		j := inputLen - i - 1

		output[j] = n
	}

	return output
}
