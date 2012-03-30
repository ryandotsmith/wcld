package main

import (
	"testing"
)

func TestGetJson(t *testing.T) {
	time, data := parseLogLine(`150 <13>1 2012-02-14T00:44:30+00:00 d.39c761b5-2e3a-4f93-9e68-2549c85650e2 app web.4 - - {"hello": "world", "time": 0.006}`)
	expected_time := "2012-02-14T00:44:30+00:00"
	expected_data := `"hello" => world, "time" => 0.006`

	if time != expected_time {
		t.Errorf("\n e(%v) \n a(%v)", expected_time, time)
	}
	if data != expected_data {
		t.Errorf("\n e(%v) \n a(%v)", expected_data, data)
	}
}

func TestGetKv(t *testing.T) {
	time, data := parseLogLine(`150 <13>1 2012-02-14T00:44:30+00:00 d.39c761b5-2e3a-4f93-9e68-2549c85650e2 app web.4 - - url="http://google.com"`)
	expected_time := "2012-02-14T00:44:30+00:00"
	expected_data := `"url" => "http://google.com"`

	if time != expected_time {
		t.Errorf("\n e(%v) \n a(%v)", expected_time, time)
	}
	if data != expected_data {
		t.Errorf("\n e(%v) \n a(%v)", expected_data, data)
	}
}

func TestToHstore(t *testing.T) {
	m := map[string]interface{}{"hello": "world"}
	actual := hstore(m)
	expected := `"hello" => world`

	if actual != expected {
		t.Errorf("\n e(%v) \n a(%v)", expected, actual)
	}
}


func TestParseOnBlank(t *testing.T) {
	_, data := parseLogLine(``)
	actual := data
	expected := ``

	if actual != expected {
		t.Errorf("\n e(%v) \n a(%v)", expected, actual)
	}
}

func TestToHstoreOnSQLLine(t *testing.T) {
	_, actual := parseLogLine(`150 <13>1 2012-02-14T00:44:30+00:00 d.39c761b5-2e3a-4f93-9e68-2549c85650e2 app web.4 - - DEBUG: (0.000863s) INSERT INTO "billable_events" ("provider_id", "rate_code_id", "entity_id", "hid", "qty", "product_name", "time", "state", "created_at") VALUES (5, 2, '40531942', '369504', 1, 'worker', '2012-02-13 18:36:30.000000+0000', 'open', '2012-02-13 18:36:49.810784+0000') RETURNING *`)
	expected := ""

	if actual != expected {
		t.Errorf("\n e(%v) \n a(%v)", expected, actual)
	}
}

func TestParseSysLog(t *testing.T) {
	match := syslogData.FindStringSubmatch(`150 <13>1 2012-02-14T00:44:30+00:00 d.39c761b5-2e3a-4f93-9e68-2549c85650e2 app web.4 - - info=true provider=3 #api_prepare_body key=value time="2012-01-01 00:00:00"`)
	actual := match[10]
	expected := `info=true provider=3 #api_prepare_body key=value time="2012-01-01 00:00:00"`

	if actual != expected {
		t.Errorf("\n e(%v) \n a(%v)", expected, actual)
	}
}

func TestParseTime(t *testing.T) {
	match := syslogData.FindStringSubmatch(`150 <13>1 2012-02-14T00:44:30+00:00 d.39c761b5-2e3a-4f93-9e68-2549c85650e2 app web.4 - - INFO: provider=3 #api_prepare_body key=value time="2012-01-01 00:00:00"`)
	actual := match[3]
	expected := "2012-02-14T00:44:30+00:00"

	if actual != expected {
		t.Errorf("\n e(%v) \n a(%v)", expected, actual)
	}
}
