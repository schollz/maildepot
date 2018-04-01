package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/mr-tron/base58/base58"
	"golang.org/x/crypto/nacl/sign"
)

var publicKey [32]byte
var privateKey [64]byte
var publicKeyString = "3YeftoEoRUk7KKUM2GZErecHpiVjXdViqv8M6pjUwXx9"
var privateKeyString = "4dDurhiCUmaAkWUAuGximzAjGTPDXZjJJzsrGq7QKAuE3pmk7JpzWRK5kWmwmxVUa9moX4DDgwu5T2D89MbWHzDB"

func init() {
	generateKeys()

	publicKeyStringBytes, err := base58.FastBase58Decoding(publicKeyString)
	if err != nil {
		panic(err)
	}
	copy(publicKey[:], publicKeyStringBytes[:32])
	privateKeyStringBytes, err := base58.FastBase58Decoding(privateKeyString)
	if err != nil {
		panic(err)
	}
	copy(privateKey[:], privateKeyStringBytes[:64])
}

func generateKeys() {
	publicKey, privateKey, _ := sign.GenerateKey(rand.Reader)
	fmt.Printf("\npublic key: %s", base58.FastBase58Encoding((*publicKey)[:]))
	fmt.Printf("\nprivate key: %s", base58.FastBase58Encoding((*privateKey)[:]))
}

func signTime() (authenticatedTime string) {
	signedMessage := sign.Sign(nil, []byte(time.Now().UTC().String()), &privateKey)
	authenticatedTime = base58.FastBase58Encoding(signedMessage)
	return
}

func authenticateSignedTime(authenticatedTime string) (actualTime string, err error) {
	signedMessage, err := base58.FastBase58Decoding(authenticatedTime)
	if err != nil {
		return
	}
	message, ok := sign.Open(nil, signedMessage, &publicKey)
	if !ok {
		err = errors.New("failed to authenticate")
		return
	}
	actualTime = string(message)
	return
}

func handlerSlash(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Time Authority API:
		
GET /now - returns the authenticated time 

GET /authenticate?now=X - returns the time given an authenticated time 

GET /public - returns the public key
`)
}

func handlerNow(w http.ResponseWriter, r *http.Request) {
	jsBytes, _ := json.Marshal(Response{
		Message: signTime(),
		Success: true,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsBytes)
}

func handlerPublic(w http.ResponseWriter, r *http.Request) {
	jsBytes, _ := json.Marshal(Response{
		Message: publicKeyString,
		Success: true,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsBytes)
}

type Response struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func handlerAuth(w http.ResponseWriter, r *http.Request) {
	var js Response
	m, _ := url.ParseQuery(r.URL.RawQuery)
	if _, ok := m["now"]; ok {
		actualTime, err := authenticateSignedTime(m["now"][0])
		if err == nil {
			js.Success = true
			js.Message = actualTime
		} else {
			js.Message = err.Error()
		}
	} else {
		js.Message = "must include ?now=X"
	}

	jsBytes, _ := json.Marshal(js)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsBytes)
}

func logHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fn(w, r)
		log.Printf("%s %s", r.URL, time.Since(start))
	}
}

func main() {
	http.HandleFunc("/now", logHandler(handlerNow))
	http.HandleFunc("/authenticate", logHandler(handlerAuth))
	http.HandleFunc("/public", logHandler(handlerPublic))
	http.HandleFunc("/", logHandler(handlerSlash))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
