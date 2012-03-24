package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	_ "github.com/bmizerany/pq"
	"log"
	"net"
	"os"
	"regexp"
)

var syslogData = regexp.MustCompile(`^(\d+) (<\d+>\d+) (\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d(\.\d+)?[\-\+]\d\d:00) ([a-zA-Z0-9\.\-]+) ([a-zA-Z0-9]+) ([a-zA-Z0-9\.]+) ([-]) ([-]) (.*)`)
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
	time, data := parseLogLine(logLine)
	if len(time) > 0 && len(data) > 0 {
		log.Printf("action=insert time=%v data=%v", time, data)
		_, err := pg.Exec("INSERT INTO log_data(time, data) VALUES ($1::hstore, $2)", time, data)
		if err != nil {
			log.Printf("error=true action=insert  message=%v", err)
		}
	}
	return
}

func parseLogLine(logLine string) (time string, data string) {
	matches := syslogData.FindStringSubmatch(logLine)

	if len(matches) >= 3 {
		time = matches[3]
	}

	if len(matches) >= 10 {
		data = hstore(getPayload(matches[10]))
	}
	return
}

func getPayload(payLoadStr string) (payLoad map[string]interface{}) {
	log.Printf("payLoadStr=%v", payLoadStr)
	if e := json.Unmarshal([]byte(payLoadStr), &payLoad); e != nil {
		log.Printf("error=%v", e)
		payLoad = map[string]interface{}{}
	}
	return
}

func hstore(data map[string]interface{}) (hstore string) {
	for k, v := range data {
		hstore += `"` + string(k) + `"` + ` => ` + `"` + v.(string) + `"`
	}
	return
}
