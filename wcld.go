package main

import (
	"bufio"
	"database/sql"
	_ "github.com/bmizerany/pq.go"
	"log"
	"net"
	"os"
	"strings"
)

var pg *sql.DB

func main() {
	var err error
	pg, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("unable to open postgres database: %v", err)
	}

	log.Println("bind tcp", os.Getenv("PORT"))
	server, err := net.Listen("tcp", "0.0.0.0:"+os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("unable to bind to tcp: %v", err)
	}

	conns := clientConns(server)
	for {
		go runD(<-conns)
	}
}

func runD(client net.Conn) {
	b := bufio.NewReadWriter(bufio.NewReader(client), bufio.NewWriter(client))
	for {
		line, err := b.ReadString('\n')
		if err != nil {
			break
		}
		handleInput(line)
	}
}

func clientConns(listener net.Listener) (ch chan net.Conn) {
	ch = make(chan net.Conn)
	go func() {
		for {
			client, err := listener.Accept()
			if err != nil {
				log.Fatalf("unable to accept tcp packet: %v", err)
			}
			log.Printf("accepted client addr: %v", client.RemoteAddr())
			ch <- client
		}
	}()
	return ch
}

func handleInput(logLine string) {
	log.Printf("handleInput logLine=%v", logLine)
	hLine := toHstore(trimKeys(logLine))
	if len(hLine) > 0 {
		log.Printf("insert into log_data data=%v", hLine)
		_, err := pg.Exec("INSERT INTO log_data (data) VALUES ($1::hstore)", hLine)
		if err != nil {
			log.Printf("insert error message=%v", err)
		}
	}
	return
}

/*
  Takes a string like "name=ryan age=25" and converts in into "name=>ryan age=>25"
*/

func toHstore(kvs string) (string) {
	return strings.Replace(kvs, "=", "=>", -1)
}

/*
  Trims a string containing key=value substrings such that
  the result is only the key=value substrings.

  i.e. string="hi #foo name=ryan age=25" would return "name=ryan age=25"
*/
func trimKeys(logLine string) (kvs string) {
	kvs = ""
	fields := strings.Fields(logLine)
	max := len(fields) - 1

	for i, elt := range fields {
		log.Printf("processing elt=%v at position=%v max=%v", elt, i, max)
		if strings.Contains(elt, "=") {
			if !(strings.HasPrefix(elt, "=") && strings.HasSuffix(elt, "=")) {
				kvs += elt
				if i != max {
					kvs += ", "
				}
			}
		}
	}
	return
}
