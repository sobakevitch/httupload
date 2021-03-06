package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	netIface  string
	localPort int
	ssl       bool
)

func logRequest(r *http.Request) {
	log.Printf("%s %s %s \"%s\" %d \"%s\"\n",
		r.RemoteAddr,
		r.Method,
		r.URL.String(),
		r.Proto,
		r.ContentLength,
		r.Header["User-Agent"])
}

func upload(w http.ResponseWriter, r *http.Request) {

	logRequest(r)
	if r.Method == "GET" {
		// GET
		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, nil)

	} else if r.Method == "POST" {
		// POST
		receivedFile, handler, err := r.FormFile("uploadfile")
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "NOK")
			return
		}
		defer receivedFile.Close()

		localFile, err := os.OpenFile(filepath.Join("./", filepath.Base(handler.Filename)), os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer localFile.Close()
		io.Copy(localFile, receivedFile)
		fmt.Fprintf(w, "OK")

	} else {
		log.Println("Unknown HTTP " + r.Method + " Method")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}

func getIPFromIfaceName(iface string) net.IP {
	var result net.IP

	if iface == "any" {
		result = net.ParseIP("0.0.0.0")
	} else {
		ief, err := net.InterfaceByName(iface)
		if err != nil {
			log.Fatal(err)
			log.Printf("Interface %s not found\n", iface)
			os.Exit(1)
		}
		addrs, err := ief.Addrs()
		if err != nil {
			log.Fatal(err)
			log.Printf("Error retrieving address for %s interface\n", iface)
			os.Exit(2)
		}
		// Retrieving the first IPv4 or IPv6 address
		for k := range addrs {
			s := strings.Split(addrs[k].String(), "/")[0]
			ip := net.ParseIP(s)
			if ip.To4() != nil { // IPv4 address
				result = ip
				break
			}
		}
		if result == nil {
			log.Printf("Error retrieving a valid IPv4 address for %s interface\n", iface)
			os.Exit(3)
		}
	}
	return result
}

func init() {
	flag.StringVar(&netIface, "i", "any", "Listen interface")
	flag.IntVar(&localPort, "p", 9090, "Listen port")
	flag.BoolVar(&ssl, "ssl", false, "SSL support")
	flag.Parse()
}

func main() {
	localAddr := getIPFromIfaceName(netIface)
	bindValue := fmt.Sprintf("%s:%d", localAddr, localPort)
	log.Printf("Listen to %s\n", bindValue)
	var err error

	http.HandleFunc("/upload", upload)
	if ssl {
		err = http.ListenAndServeTLS(bindValue, "server.crt", "server.key", nil)
	} else {
		err = http.ListenAndServe(bindValue, nil)
	}
	log.Fatal(err)
}
