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
	logData := toHstore(trimKeys(logLine))
	logTime := parseTime(logLine)
	if len(logData) > 0 {
		log.Printf("action=insert logData=%v logTime=%v", logData, logTime)
		_, err := pg.Exec("INSERT INTO log_data (data, time) VALUES ($1::hstore, $2)", logData, logTime)
		if err != nil {
			log.Printf("error=true action=insert  message=%v", err)
		}
	}
	return
}

func parseTime(logLine string) (time string) {
	t, _ := regexp.Compile(`(\d\d\d\d)(-)?(\d\d)(-)?(\d\d)(T)?(\d\d)(:)?(\d\d)(:)?(\d\d)(\.\d+)?(Z|([+-])(\d\d)(:)?(\d\d))`)
	time = t.FindString(logLine)
	log.Printf("action=parse_time t=%v", time)
	return
}

func toHstore(kvs string) string {
	return strings.Replace(kvs, "=", "=>", -1)
}

func trimKeys(logLine string) (kvs string) {
	kv, _ := regexp.Compile("([a-z0-9_.-]+)=([a-z0-9_.-]+)")
	pairs := kv.FindAllString(logLine, -1)
	max := len(pairs) - 1

	for i, elt := range pairs {
		log.Printf("action=process_element elt=%v position=%v max=%v", elt, i, max)
		kvs += elt
		if i != max {
			kvs += ", "
		}
	}
	return
}
