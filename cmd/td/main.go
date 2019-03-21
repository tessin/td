package main

import (
	"log"
	"os"

	"github.com/tessin/td"
)

func main() {
	argv := os.Args[1:]

	if 0 < len(argv) {
		switch argv[0] {
		case "pack":
			err := td.Pack()
			if err != nil {
				log.Fatal(err)
			}
			break
		case "deploy":
			err := td.Pack()
			if err != nil {
				log.Fatal(err)
			}
			err = td.Deploy(argv[1], argv[2])
			if err != nil {
				log.Fatal(err)
			}
			break
		}
	}

	// ctx, err := az.GetContext(az.ResourceManager)
	// if err != nil {
	// 	log.Println("cannot acquire Azure context")
	// 	log.Fatal(err)
	// }
	// log.Println("subscription", ctx.Subscription.Id, ctx.Subscription.Name)

	// list, err := ctx.GetWebSiteList("TCM")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println(list)
}
