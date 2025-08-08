// main.go
package main

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hoomandigital/terraform-provider-nosana/nosana"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: nosana.Provider,
	})
	log.Println("Nosana Terraform Provider started.")
}