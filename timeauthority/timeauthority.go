package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/mr-tron/base58/base58"
	"golang.org/x/crypto/nacl/sign"
)

var publicKey [32]byte
var privateKey [64]byte

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
`)
}

func handlerNow(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", signTime())
}

func handlerAuth(w http.ResponseWriter, r *http.Request) {
	m, _ := url.ParseQuery(r.URL.RawQuery)
	if _, ok := m["now"]; ok {
		actualTime, err := authenticateSignedTime(m["now"][0])
		if err == nil {
			fmt.Fprint(w, actualTime)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err.Error())
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "see / for usage")
	}
}

func logHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fn(w, r)
		log.Printf("%s %s", r.URL, time.Since(start))
	}
}

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "port to listen on")
	flag.Parse()
	type Keys struct {
		Public  string `json:"public"`
		Private string `json:"private"`
	}
	var k Keys
	if _, err := os.Stat("keys.json"); os.IsNotExist(err) {
		log.Println("Generating new keys...")
		publicKey, privateKey, _ := sign.GenerateKey(rand.Reader)
		k.Public = base58.FastBase58Encoding((*publicKey)[:])
		k.Private = base58.FastBase58Encoding((*privateKey)[:])
		bKeys, _ := json.Marshal(k)
		err := ioutil.WriteFile("keys.json", bKeys, 0644)
		if err != nil {
			panic(err)
		}
	} else {
		bKeys, err := ioutil.ReadFile("keys.json")
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(bKeys, &k)
		if err != nil {
			panic(err)
		}
	}

	publicKeyStringBytes, err := base58.FastBase58Decoding(k.Public)
	if err != nil {
		panic(err)
	}
	copy(publicKey[:], publicKeyStringBytes[:32])
	privateKeyStringBytes, err := base58.FastBase58Decoding(k.Private)
	if err != nil {
		panic(err)
	}
	copy(privateKey[:], privateKeyStringBytes[:64])

	http.HandleFunc("/now", logHandler(handlerNow))
	http.HandleFunc("/authenticate", logHandler(handlerAuth))
	http.HandleFunc("/", logHandler(handlerSlash))
	fmt.Printf("listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
