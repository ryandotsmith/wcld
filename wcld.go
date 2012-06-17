package main

import (
	"bufio"
	"database/sql"
	"flag"
	"github.com/bmizerany/pq"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

var checkpoint = flag.Int("checkpoint", 1, "1 for max durability, 1000 for max throughput")
var pg *sql.DB

var LineRe = regexp.MustCompile(`\d+ \<\d+\>1 \d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d\+00:00 d\.[a-z0-9-]+ ([a-z0-9\-\_\.]+) ([a-z0-9\-\_\.]+) \- \- (.*)`)
var AttrsRe = regexp.MustCompile(`( *)([a-zA-Z0-9\_\-\.]+)=?(([a-zA-Z0-9\.\-\_\.]+)|("([^\"]+)"))?`)

func main() {
	flag.parse()
	cs, err := pq.ParseURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("unable to parse database url")
		os.Exit(1)
	}

	pg, err = sql.Open("postgres", cs)
	if err != nil {
		log.Println("unable to connect to database")
		os.Exit(1)
	}

	log.Println("bind tcp", os.Getenv("PORT"))
	server, err := net.Listen("tcp", "0.0.0.0:"+os.Getenv("PORT"))
	if err != nil {
		log.Println("unable to bind tcp")
		os.Exit(1)
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
				log.Printf("error=true action=tcp_accept message=%v\n", err)
			}
			log.Printf("action=tcp_accept remote= %v\n", client.RemoteAddr())
			ch <- client
		}
	}()
	return ch
}

func readData(client net.Conn) {
	b := bufio.NewReadWriter(bufio.NewReader(client), bufio.NewWriter(client))
	i := 0
	var err error
	var tx *sql.Tx
	for {
		if i == 0 {
			tx, err = pg.Begin()
			if err != nil {
				log.Printf("error=true action=begin message=%v", err)
			}
			i += 1
		} else if i == (checkpoint + 1) {
			//checkpoint is set by flag
			// we inc checkpoint for the case when it is set to 1
			err = tx.Commit()
			if err != nil {
				log.Printf("error=true action=commit message=%v", err)
			}
			log.Printf("action=commit")
			i = 0
		} else {
			line, err := b.ReadString('\n')
			if err != nil {
				break
			}
			handleInput(*tx, line)
			i += 1
		}
	}
}

func handleInput(tx sql.Tx, logLine string) {
	data := hstore(parse(logLine))
	if len(data) > 0 {
		_, err := tx.Exec("INSERT INTO log_data (data, time) VALUES ($1::hstore, now())", data)
		if err != nil {
			log.Printf("error=true action=insert  \n message=%v \n data=%v\n", err, data)
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
	logLine = strings.Trim(logLine, "\n")
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
