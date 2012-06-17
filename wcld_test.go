package main

import (
	"testing"
)

func TestParse(t *testing.T) {
	m := parse(`150 <13>1 2012-02-14T00:44:30+00:00 d.39c761b5-2e3a-4f93-9e68-2549c85650e2 app web.4 - - test name=ryan age=25 _ssn=123 description= fav-quote="oh=hai"`)

	if m["name"] != "ryan" {
		t.Errorf("name != ryan")
	}
	if m["fav-quote"] != `"oh=hai"` {
		t.Errorf(`fav-quote!= "oh=hai"`)
	}
	if m["_ssn"] != "123" {
		t.Errorf("ssn != 123")
	}
}
