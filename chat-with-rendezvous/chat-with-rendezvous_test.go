package main

import (
	"testing"
	"time"
)

//"Chat with rendezvous" will poll for input forever, here just run the main logic to see if there are any runtime errors
//FIXME: As of the time of writing, Chat with rendezvous doesn't work properly, yet this test does not fail.
func TestMain(t *testing.M) {
	timer1 := time.NewTimer(2 * time.Second)
	go main()
	<-timer1.C
}
