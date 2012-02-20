package main

import (
	"testing"
)

func TestTrimKeys(t *testing.T) {
	actual := trimKeys(`#test name=ryan age=25 height-feet=6 height-inches=5 weight_lbs=210 _ssn=123 fav_quote="oh hai"`)
	expected := `name=ryan, age=25, height-feet=6, height-inches=5, weight_lbs=210, _ssn=123, fav_quote="oh hai"`

	if actual != expected {
		t.Errorf("expected %v actual: %v", expected, actual)
	}
}

func TestToHstore(t *testing.T) {
	trimed := trimKeys("#test name=ryan age=25")
	actual := toHstore(trimed)
	expected := "name=>ryan, age=>25"

	if actual != expected {
		t.Errorf("expected %v actual: %v", expected, actual)
	}
}

func TestToHstoreOnRouterLine(t *testing.T) {
	trimed := trimKeys("PUT shushu.herokuapp.com/resources/584093/billable_events/40531647 dyno=web.3 queue=0 wait=0ms service=52ms status=201 bytes=239")
	actual := toHstore(trimed)
	expected := "dyno=>web.3, queue=>0, wait=>0ms, service=>52ms, status=>201, bytes=>239"

	if actual != expected {
		t.Errorf("expected %v actual: %v", expected, actual)
	}
}

func TestToHstoreOnSQLLine(t *testing.T) {
	trimed := trimKeys(`DEBUG: (0.000863s) INSERT INTO "billable_events" ("provider_id", "rate_code_id", "entity_id", "hid", "qty", "product_name", "time", "state", "created_at") VALUES (5, 2, '40531942', '369504', 1, 'worker', '2012-02-13 18:36:30.000000+0000', 'open', '2012-02-13 18:36:49.810784+0000') RETURNING *`)
	actual := toHstore(trimed)
	expected := ""

	if actual != expected {
		t.Errorf("expected %v actual: %v", expected, actual)
	}
}

func TestParseTime(t *testing.T) {
	actual := parseTime("<13>1 2012-02-14T00:44:30+00:00 d.39c761b5-2e3a-4f93-9e68-2549c85650e2 app web.4 - - INFO: provider=3 #api_prepare_body key=value")
	expected := "2012-02-14T00:44:30+00:00"

	if actual != expected {
		t.Errorf("expected %v actual: %v", expected, actual)
	}
}
