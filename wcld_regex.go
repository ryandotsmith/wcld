package main

import (
	"regexp"
	"os"
)

var AcceptPattern = regexp.MustCompile( os.Getenv("ACCEPT_PATTERN") )

var KeyPattern = regexp.MustCompile(`\w+=`)

var KvSig = regexp.MustCompile(`([a-zA-Z0-9\.\_\-\:\/])=([a-zA-Z0-9\.\_\-\:\/\"\'])`)

var KvData = regexp.MustCompile(`([a-zA-Z0-9\.\_\-\:\/]+)(=?)("[^"]+"|'[^']+'|[a-zA-Z0-9\.\_\-\:\/]*)`)

var SyslogData = regexp.MustCompile(`^(\d+) (<\d+>\d+) (\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d(\.\d+)?[\-\+]\d\d:00) ([a-zA-Z0-9\.\-]+) ([a-zA-Z0-9]+) ([a-zA-Z0-9\.]+) ([-]) ([-]) (.*)`)
