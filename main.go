// main.go
package main

import (
	"log"

	"github.com/HoomanDigital/TerraformProvider-Nosana/nosana"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: nosana.Provider,
	})
	log.Println("Nosana Terraform Provider started.")
}
