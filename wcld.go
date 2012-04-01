package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ryandotsmith/lscan"
	"github.com/ryandotsmith/pq"
	"flag"
	"net"
	"os"
	"regexp"
	"strings"
)

var sType *string = flag.String("f", "", "force parser to use json or kv")

var syslogData = regexp.MustCompile(`^(\d+) (<\d+>\d+) (\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d(\.\d+)?[\-\+]\d\d:00) ([a-zA-Z0-9\.\-]+) ([a-zA-Z0-9]+) ([a-zA-Z0-9\.]+) ([-]) ([-]) (.*)`)
var pg *sql.DB

func main() {
	flag.Parse()

	cs, err := pq.ParseURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println("unable to parse database url")
		os.Exit(1)
	}

	pg, err = sql.Open("postgres", cs)
	if err != nil {
		fmt.Println("unable to connect to database")
		os.Exit(1)
	}

	fmt.Println("bind tcp", os.Getenv("PORT"))
	server, err := net.Listen("tcp", "0.0.0.0:"+os.Getenv("PORT"))
	if err != nil {
		fmt.Println("unable to bind tcp")
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
				fmt.Printf("error=true action=tcp_accept message=%v\n", err)
			}
			fmt.Printf("action=tcp_accept remote= %v\n", client.RemoteAddr())
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
	time, data := parseLogLine(logLine)
	if len(time) > 0 && len(data) > 0 {
		_, err := pg.Exec("INSERT INTO log_data(time, data) VALUES ($1, $2::hstore)", time, data)
		if err != nil {
			fmt.Printf("error=true action=insert  \n message=%v \n data=%v\n", err, data)
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
		sMatch := matches[10]
		switch *sType {
		case "json":
			if d := getJson(sMatch); len(d) > 0 {
				data = hstore(d)
			}
		case "kv":
			if d := getKv(sMatch); len(d) > 0 {
				data = hstore(d)
			}
		default:
			if d := getJson(sMatch); len(d) > 0 {
				data = hstore(d)
			} else if d := getKv(sMatch); len(d) > 0 {
				data = hstore(d)
			}
		}
	}
	return
}

func hstore(data map[string]interface{}) (hstore string) {
	max := len(data)
	i := 0
	for k, v := range data {
		i += 1
		hstore += `"` + string(k) + `"` + ` => ` + fmt.Sprintf("%v", v)
		if i != max {
			hstore += ", "
		}
	}
	return
}

func getJson(payLoadStr string) (payLoad map[string]interface{}) {
	if e := json.Unmarshal([]byte(payLoadStr), &payLoad); e != nil {
		payLoad = map[string]interface{}{}
	}
	return
}

func getKv(payLoadStr string) (payLoad map[string]interface{}) {
	r := strings.NewReader(payLoadStr)
	payLoad = lscan.Parse(r)
	return
}
