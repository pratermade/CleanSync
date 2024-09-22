package main

import "testing"

func TestSync(t *testing.T) {
	err := sync("pratermade-gotest", "C:\\Users\\pratersm\\Documents\\WebSites\\yci-www")
	if err != nil {
		t.Fatal(err)
	}

}
