package main

import (
	"bufio"
	"database/sql"
	"github.com/bmizerany/pq"
	"log"
	"net"
	"os"
	"regexp"
)

var pg *sql.DB

var LineRe = regexp.MustCompile(`\d+ \<\d+\>1 \d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d\+00:00 d\.[a-z0-9\-]+ ([a-z0-9\-\_\.]+) ([a-z0-9\-\_\.]+) \- \- (.*)$`)
var AttrsRe = regexp.MustCompile(`( *)([a-zA-Z0-9\_\-\.]+)=?(([a-zA-Z0-9\.\-\_\.]+)|("([^\"]+)"))?`)

func main() {
	cs, err := pq.ParseURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("unable to parse database url")
		os.Exit(1)
	}

	pg, err = sql.Open("postgres", cs)
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
	data := hstore(parse(logLine))
	if len(data) > 0 {
		_, err := pg.Exec("INSERT INTO log_data (data, time) VALUES ($1::hstore, now())", data)
		if err != nil {
			log.Printf("error=true action=insert  message=%v", err)
		}
	}
	return
}

func hstore(m map[string]string) (s string) {
	i := 0
	max := len(m)
	for k, v := range m {
		s += k + `=>` + v
		i += 1
		if i != max {
			s += ", "
		}
	}
	return
}

func parse(logLine string) map[string]string {
	kvs := make(map[string]string)
	data := LineRe.FindStringSubmatch(logLine)

	if len(data) > 0 {
		d := data[3]
		words := AttrsRe.FindAllStringSubmatch(d, -1)
		for _, match := range words {
			k := match[2]
			v1 := match[3]
			v2 := match[5]
			if len(v1) != 0 {
				kvs[k] = v1
			} else if len(v2) != 0 {
				kvs[k] = v2
			} else {
				kvs[k] = "true"
			}
		}

	}
	return kvs
}
