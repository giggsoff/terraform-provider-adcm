package main

import (
	"context"
	"fmt"

	"github.com/giggsoff/terraform-provider-adcm/adcm"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	err := providerserver.Serve(context.Background(), adcm.New,
		providerserver.ServeOpts{Address: "github.com/giggsoff/adcm"},
	)
	if err != nil {
		fmt.Println(err)
	}
}
