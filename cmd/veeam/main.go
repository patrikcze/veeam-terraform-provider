package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/patrikcze/terraform-provider-veeam/internal"
)

func main() {
	providerserver.Serve(context.Background(), internal.New("dev"), providerserver.ServeOpts{
		Address: "registry.terraform.io/patrikcze/veeam",
	})
}
