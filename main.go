package main

import (
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"
)

var (
	rootCAFilename = filepath.Join("pki", "rootCA.crt")
	certPool       = x509.NewCertPool()
	serverFlag     = flag.Bool("s", false, "run as server")
	clientFlag     = flag.Bool("c", false, "run as client")
	addressFlag    = flag.String("a", "0.0.0.0:3000", "address to listen (server), or send message to (client)")
	pkiFlag        = flag.Bool("k", false, "(re)generate private keys and certificates")
)

// For TLS client server, see
// - https://gist.github.com/spikebike/2232102
// - https://gist.github.com/ncw/9253562
// For JSON marshalling and unmarshalling, see
// - http://choly.ca/post/go-json-marshalling
// For Database interactions, see
// - https://astaxie.gitbooks.io/build-web-application-with-golang/en/05.2.html

func main() {
	flag.Parse()

	if *pkiFlag {
		log.Println("(re)generating private keys and certificates")
		createPKI()
	}

	data, err := ioutil.ReadFile(rootCAFilename)
	if err != nil {
		log.Fatalln(err)
	}
	if !certPool.AppendCertsFromPEM(data) {
		log.Fatalf("failed to parse rootCA certificate '%s'\n", rootCAFilename)
	}

	if *serverFlag {
		runAsServer()
	} else if *clientFlag {
		runAsClient()
	}
}
