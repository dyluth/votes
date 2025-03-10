package main

import (
	"fmt"

	"github.com/dyluth/votes/publicwhip"
)

func main() {

	publicwhip.SetupMPs()
	for n, v := range publicwhip.AllMPs {
		for n2, v2 := range publicwhip.Policies {
			position, err := publicwhip.GetMPPolicyPosition(v, v2)
			if err != nil {
				panic(err)
			}
			fmt.Printf("\n\n%v %v: %v\n", n, position, n2)
			return
		}
	}
}
