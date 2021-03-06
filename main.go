package main // import "github.com/chmike/go-dmon"

import (
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/pkg/profile"
)

var (
	rootCAFilename = filepath.Join("pki", "rootCA.crt")
	certPool       = x509.NewCertPool()
	serverFlag     = flag.Bool("s", false, "run as server")
	clientFlag     = flag.Bool("c", false, "run as client")
	addressFlag    = flag.String("a", "127.0.0.1:3000", "server: listen address, client: message destination")
	pkiFlag        = flag.Bool("k", false, "(re)generate private keys and certificates")
	dbFlag         = flag.Bool("db", false, "store monitoring messages in database")
	tlsFlag        = flag.Bool("tls", false, "use TLS connection (default tcp)")
	jsonFlag       = flag.Bool("json", false, "use json encoding (default binary)")
	cpuFlag        = flag.Bool("cpu", false, "enable CPU profiling")
	periodFlag     = flag.Int("p", 5, "stat display period in seconds")
	dbFlushFlag    = flag.Int("dbp", 1000, "database flush period in milliseconds")
	dbBufLenFlag   = flag.Int("dbl", 10, "database buffer length")
	msgFlag        = flag.Bool("m", false, "display received messages")
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

	if *cpuFlag {
		defer profile.Start().Stop()
	}

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

	switch {
	case *serverFlag:
		runAsServer()
	case *clientFlag:
		runAsClient()
	default:
		flag.Usage()
		log.Fatalf("need either to run as server or as client")
	}
}
