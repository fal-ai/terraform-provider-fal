package main

import (
	"context"
	"flag"
	"log"

	"github.com/fal-ai/terraform-provider-fal/fal"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is set via ldflags during build
var version = "dev"

// Generate the Terraform provider documentation using `tfplugindocs`:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(
		context.Background(),
		fal.New(version),
		providerserver.ServeOpts{
			Address: "registry.terraform.io/fal-ai/fal",
			Debug:   debugMode,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
