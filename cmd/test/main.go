package main

import (
	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"github.com/abuse-mesh/abuse-mesh-go/pkg/client"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	client := client.NewAbuseMeshClient()

	spew.Dump(client.GetNode(&abusemesh.GetNodeRequest{}))

	client.Close()
}
