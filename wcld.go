package main

import (
	"bufio"
	"database/sql"
	_ "github.com/bmizerany/pq.go"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

var pg *sql.DB

var kvSig = regexp.MustCompile(`([a-zA-Z0-9\.\_\-\:\/])=([a-zA-Z0-9\.\_\-\:\/\"])`)
var kvData = regexp.MustCompile(`([a-zA-Z0-9\.\_\-\:\/]+)(=?)("[^"]+"|'[^']+'|[a-zA-Z0-9\.\_\-\:\/]*)`)
var syslogData = regexp.MustCompile(`^(\d+) (<\d+>\d+) (\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d(\.\d+)?[\-\+]\d\d:00) ([a-zA-Z0-9\.\-]+) ([a-zA-Z0-9]+) ([a-zA-Z0-9\.]+) ([-]) ([-]) (.*)`)

func main() {
	var err error
	pg, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("error=true action=db_conn message=%v", err)
	}
	log.Println("bind tcp", os.Getenv("PORT"))
	server, err := net.Listen("tcp", "0.0.0.0:"+os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("error=true action=net_listen message=%v", err)
	}
	conns := clientConns(server)
	for {
		go readData(<-conns)
	}
}

func clientConns(listener net.Listener) (ch chan net.Conn) {
	ch = make(chan net.Conn)
	go func() {
		for {
			client, err := listener.Accept()
			if err != nil {
				log.Printf("error=true action=tcp_accept message=%v", err)
			}
			log.Printf("action=tcp_accept remote= %v", client.RemoteAddr())
			ch <- client
		}
	}()
	return ch
}

func readData(client net.Conn) {
	b := bufio.NewReadWriter(bufio.NewReader(client), bufio.NewWriter(client))
	for {
		line, err := b.ReadString('\n')
		if err != nil {
			break
		}
		handleInput(line)
	}
}

func handleInput(logLine string) {
	log.Printf("action=handleInput logLine=%v", logLine)
	logTime := syslogData.FindStringSubmatch(logLine)[3]
	logData := toHstore(logLine)
	if len(logData) > 0 {
		log.Printf("action=insert logData=%v logTime=%v", logData, logTime)
		_, err := pg.Exec("INSERT INTO log_data (data, time) VALUES ($1::hstore, $2)", logData, logTime)
		if err != nil {
			log.Printf("error=true action=insert  message=%v", err)
		}
	}
	return
}

func toHstore(logLine string) string {
	message := syslogData.FindStringSubmatch(logLine)[10]
	words := kvData.FindAllString(message, -1)
	max := len(words) - 1
	hasSig := kvSig.FindAllString(message, -1)
	kvs := ""

	if hasSig != nil {
		for i, elt := range words {
			if kvSig.MatchString(elt) {
				kvs += elt
			} else if m, _ := regexp.MatchString(`\w+=`, elt); m == true {
				kvs += elt + `""`
			} else {
				kvs += `"` + elt + `"` + "=true"
			}
			if i != max {
				kvs += ", "
			}
		}
	} else {
		message = strings.Replace(message, `"`, `'`, -1)
		kvs = `message="` + message + `"`
	}
	return strings.Replace(kvs, "=", "=>", -1)
}
