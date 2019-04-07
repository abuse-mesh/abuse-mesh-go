package cmd

//This file contains all commands related to the node we are connected to

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/abuse-mesh/abuse-mesh-go/pkg/adminapi"
	"github.com/abuse-mesh/abuse-mesh-go/pkg/adminapiclient"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

func init() {
	// ./abusemesh get node
	getCmd.AddCommand(getNodeCommand)
}

//Get information about the node we are connected to
var getNodeCommand = &cobra.Command{
	Use:   "node",
	Short: "Get all information about the node",
	Run: func(cmd *cobra.Command, args []string) {
		client := adminapiclient.NewAbuseMeshAdminClient()

		node, err := client.GetNode(&adminapi.GetNodeRequest{})
		if err != nil {
			exitWithGrpcError(err)
		}

		printToStdout(node, func(object interface{}) string {
			buf := &bytes.Buffer{}
			tabWriter := tabwriter.NewWriter(buf, 0, 0, 3, ' ', 0)

			fmt.Fprintln(buf, "==================== General details ====================")

			fmt.Fprintf(tabWriter, "UUID:\t%s\n", node.GetUuid().GetUuid())
			fmt.Fprintf(tabWriter, "Protocol version:\t%s\n", node.GetProtocolVersion())
			fmt.Fprintf(tabWriter, "IP address:\t%s\n", node.GetIpAddress().Address)

			tabWriter.Flush()

			fmt.Fprintln(buf, "\n==================== Contact details ====================")

			fmt.Fprintf(tabWriter, "Organization:\t%s\n", node.ContactDetails.OrganizationName)
			fmt.Fprintf(tabWriter, "Email:\t%s\n", node.ContactDetails.EmailAddress)
			fmt.Fprintf(tabWriter, "Phone number:\t%s\n", node.ContactDetails.PhoneNumber)
			fmt.Fprintf(tabWriter, "Physical address:\t%s\n", node.ContactDetails.PhysicalAddress)

			tabWriter.Flush()

			fmt.Fprintln(buf, "\n==================== Contact persons ====================")

			for _, contactPerson := range node.ContactDetails.ContactPersons {

				fmt.Fprintf(tabWriter, "Name:\t%s %s %s\n", contactPerson.FirstName, contactPerson.MiddleName, contactPerson.LastName)
				fmt.Fprintf(tabWriter, "Job title:\t%s\n", contactPerson.JobTitle)
				fmt.Fprintf(tabWriter, "Email:\t%s\n", contactPerson.EmailAddress)
				fmt.Fprintf(tabWriter, "Phone number:\t%s\n", contactPerson.PhoneNumber)

				tabWriter.Flush()

				fmt.Fprintln(buf, strings.Repeat("-", 57))
			}

			fmt.Fprintln(buf, "\n==================== PGP details ====================")

			bytesReader := bytes.NewReader(node.PgpEntity.GetPgpPackets())

			pgpPacketReader := packet.NewReader(bytesReader)

			pgpEntity, err := openpgp.ReadEntity(pgpPacketReader)
			if err != nil {
				exitWithError(errors.Wrap(err, "Unable to read PGP entity of node"))
			}

			fmt.Fprintf(tabWriter, "Key ID:\t%s\n", pgpEntity.PrimaryKey.KeyIdString())
			fmt.Fprintf(tabWriter, "Key fingerprint:\t%X %X %X %X %X %X %X %X %X %X\n",
				pgpEntity.PrimaryKey.Fingerprint[0:2],
				pgpEntity.PrimaryKey.Fingerprint[2:4],
				pgpEntity.PrimaryKey.Fingerprint[4:6],
				pgpEntity.PrimaryKey.Fingerprint[6:8],
				pgpEntity.PrimaryKey.Fingerprint[8:10],
				pgpEntity.PrimaryKey.Fingerprint[10:12],
				pgpEntity.PrimaryKey.Fingerprint[12:14],
				pgpEntity.PrimaryKey.Fingerprint[14:16],
				pgpEntity.PrimaryKey.Fingerprint[16:18],
				pgpEntity.PrimaryKey.Fingerprint[18:20])

			fmt.Fprintf(tabWriter, "Created:\t%s\n", pgpEntity.PrimaryKey.CreationTime)

			tabWriter.Flush()

			fmt.Fprintln(buf, "\n-------------------- Identities ---------------------")

			for name := range pgpEntity.Identities {
				fmt.Fprintf(tabWriter, "Name:\t%s\n", name)

				tabWriter.Flush()

				fmt.Fprintln(buf, strings.Repeat("-", 57))
			}
			// fmt.Fprintf(tabWriter, ":\t%s\n", pgpEntity)
			// fmt.Fprintf(tabWriter, ":\t%s\n", pgpEntity)

			tabWriter.Flush()

			return buf.String()
		})
	},
}
