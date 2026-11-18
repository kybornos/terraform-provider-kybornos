package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/kybornos/terraform-provider-kybornos/internal/provider"
)

func main() {
	debug := flag.Bool("debug", false, "Enable plugin debugging")
	flag.Parse()
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.opentofu.org/kybornos/kybornos", Debug: *debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
