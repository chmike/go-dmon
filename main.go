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
	addressFlag    = flag.String("a", "127.0.0.1:3000", "address to listen (server), or send message to (client)")
	pkiFlag        = flag.Bool("k", false, "(re)generate private keys and certificates")
	dbFlag         = flag.Bool("db", false, "store monitoring messages in database")
	tlsFlag        = flag.Bool("tls", false, "use TLS connection")
	msgFlag        = flag.String("msg", "binary", "message recv/send protocol (json, binary)")
	cpuFlag        = flag.Bool("cpu", false, "enable CPU profiling")
	periodFlag     = flag.Int("p", 5, "period of stat display (seconds)")
	bufPeriodFlag  = flag.Int("bp", 1000, "period of bufferized send (msec)")
	bufLenFlag     = flag.Int("bl", 4096, "size of send buffer (0 = none)")
)

const (
	// JSON encode messages
	JSON = iota
	// BINARY encode messages
	BINARY
)

var msgCodec = BINARY

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

	switch *msgFlag {
	case "binary":
		msgCodec = BINARY
	case "json":
		msgCodec = JSON
	default:
		flag.Usage()
		log.Fatalf("invalid msg flag value (%s). want json or binary.", *msgFlag)
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
