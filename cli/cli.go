package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const addr = ":5000"

func main() {
	//! Doing for testing purpose

	rand.Seed(time.Now().UnixNano())
	// uid := rand.Int()
	// certBytes := getCertificate()
	// cli := app.New(certBytes, uid)
	// hs := httpServer{cli}
	hs := testhttpServer{}

	router := mux.NewRouter()
	get := router.Methods("GET").Subrouter()
	del := router.Methods("DELETE").Subrouter()
	post := router.Methods("POST").Subrouter()
	get.HandleFunc("/tunnels", hs.ListTunnels)
	del.HandleFunc("/tunnel", hs.DisconnectTunnel)
	post.HandleFunc("/tunnel/local", hs.CreateLocalTunnel)
	post.HandleFunc("/tunnel/remote", hs.CreateRemoteTunnel)
	router.Use(mux.CORSMethodMiddleware(router))

	log.Print("Listening for request at ", addr, "...\n")
	http.ListenAndServe(addr, router)
}

func getCertificate() []byte {
	h, err := os.UserHomeDir()
	if err != nil {
		log.Print("Couldn't find the user home directory: ", err)
		return nil
	}
	filePath := strings.Join([]string{h, ".shree.cert.pem"}, "/")

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Print("Couldn't read certificate from local storage: ", err)
	}

	return content

}
