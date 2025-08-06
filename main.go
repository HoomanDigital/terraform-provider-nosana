// main.go
package main

import (
	// "context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
	log.Println("Nosana Terraform Provider started.")
}