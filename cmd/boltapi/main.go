package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/boltdb/bolt"

	"github.com/marconi/boltapi"
)

var (
	dbpath = flag.String("dbpath", "", "Path to bolt database")
	port   = flag.Int("port", 8080, "Port to listen to")
)

func main() {
	flag.Parse()
	if strings.TrimSpace(*dbpath) == "" {
		log.Fatal("-dbpath param is required")
	}

	db, err := bolt.Open(*dbpath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	log.Fatal(boltapi.Serve(db, *port))
}
