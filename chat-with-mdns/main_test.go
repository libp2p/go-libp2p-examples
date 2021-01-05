package main

import(
	"testing"
	"time"
)

//"Chat with mdns" will poll for input forever, here just run the main logic to see if there are any runtime errors
func TestMain(t *testing.M){
	timer1 := time.NewTimer(2 * time.Second)
	go main()
	<-timer1.C
}