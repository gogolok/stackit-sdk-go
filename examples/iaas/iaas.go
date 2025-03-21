package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/stackitcloud/stackit-sdk-go/core/config"
	"github.com/stackitcloud/stackit-sdk-go/core/runtime"
	"github.com/stackitcloud/stackit-sdk-go/core/utils"
	"github.com/stackitcloud/stackit-sdk-go/services/iaas"
	"github.com/stackitcloud/stackit-sdk-go/services/iaas/wait"
)

func main() {
	// Specify the organization ID and project ID
	organizationId := "ORGANIZATION_ID"
	projectId := "PROJECT_ID"

	// Create a new API client, that uses default authentication and configuration
	iaasClient, err := iaas.NewAPIClient(
		config.WithRegion("eu01"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Creating API client: %v\n", err)
		os.Exit(1)
	}

	// List the network areas for your organization
	var areas *iaas.NetworkAreaListResponse //nolint:golint // transparency on data model naming
	areas, err = iaasClient.ListNetworkAreas(context.Background(), organizationId).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when calling `ListNetworkAreas`: %v\n", err)
	} else {
		fmt.Printf("[IaaS API] Number of network areas: %v\n", len(*areas.Items))
	}

	// Create a network area
	createNetworkAreaPayload := iaas.CreateNetworkAreaPayload{
		Name: utils.Ptr("example-network-area"),
		AddressFamily: &iaas.CreateAreaAddressFamily{
			Ipv4: &iaas.CreateAreaIPv4{
				DefaultPrefixLen: utils.Ptr(int64(25)),
				MaxPrefixLen:     utils.Ptr(int64(29)),
				MinPrefixLen:     utils.Ptr(int64(24)),
				NetworkRanges: &[]iaas.NetworkRange{
					{
						Prefix: utils.Ptr("1.2.3.4/24"),
					},
				},
				TransferNetwork: utils.Ptr("1.2.4.5/24"),
			},
		},
	}
	area, err := iaasClient.CreateNetworkArea(context.Background(), organizationId).CreateNetworkAreaPayload(createNetworkAreaPayload).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when calling `CreateNetworkAreas`: %v\n", err)
	} else {
		fmt.Printf("[IaaS API] Triggered creation of network area with ID %q.\n", *area.AreaId)
	}

	// Wait for creation of the network area
	_, err = wait.CreateNetworkAreaWaitHandler(context.Background(), iaasClient, organizationId, *area.AreaId).WaitWithContext(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when waiting for creation: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[IaaS API] Network area %q has been successfully created.\n", *area.AreaId)

	// Update a network area
	updateNetworkAreaPayload := iaas.PartialUpdateNetworkAreaPayload{
		Name: utils.Ptr(*area.Name + "-renamed"),
	}
	updatedArea, err := iaasClient.PartialUpdateNetworkArea(context.Background(), organizationId, *area.AreaId).PartialUpdateNetworkAreaPayload(updateNetworkAreaPayload).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when calling `PartialUpdateNetworkArea`: %v\n", err)
	} else {
		fmt.Printf("[IaaS API] Triggered update of network area with ID %q.\n", *updatedArea.AreaId)
	}

	// Wait for update of the network area
	_, err = wait.UpdateNetworkAreaWaitHandler(context.Background(), iaasClient, organizationId, *updatedArea.AreaId).WaitWithContext(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when waiting for update: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[IaaS API] Network area %q has been successfully updated.\n", *updatedArea.AreaId)

	// Delete a network area
	err = iaasClient.DeleteNetworkArea(context.Background(), organizationId, *updatedArea.AreaId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when calling `DeleteNetworkArea`: %v\n", err)
	} else {
		fmt.Printf("[IaaS API] Triggered deletion of network area with ID %q.\n", *updatedArea.AreaId)
	}

	// Wait for deletion of the network area
	_, err = wait.DeleteNetworkAreaWaitHandler(context.Background(), iaasClient, organizationId, *updatedArea.AreaId).WaitWithContext(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when waiting for deletion: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[IaaS API] Network area %q has been successfully deleted.\n", *updatedArea.AreaId)

	// Create a network
	createNetworkPayload := iaas.CreateNetworkPayload{
		Name: utils.Ptr("example-network"),
		AddressFamily: &iaas.CreateNetworkAddressFamily{
			Ipv4: &iaas.CreateNetworkIPv4{
				PrefixLength: utils.Ptr(int64(24)),
				Nameservers:  &[]string{"1.2.3.4"},
			},
		},
	}

	var httpResp *http.Response
	ctxWithHTTPResp := runtime.WithCaptureHTTPResponse(context.Background(), &httpResp)
	err = iaasClient.CreateNetwork(ctxWithHTTPResp, projectId).CreateNetworkPayload(createNetworkPayload).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when calling `CreateNetwork`: %v\n", err)
		os.Exit(1)
	}

	network, err := wait.CreateNetworkWaitHandler(context.Background(), iaasClient, projectId, httpResp.Header.Get("x-request-id")).WaitWithContext(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when waiting for creation: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[IaaS API] Network has been successfully created.\n")

	// Update a network
	updateNetworkPayload := iaas.PartialUpdateNetworkPayload{
		Name: utils.Ptr("example-network-test-renamed"),
	}

	err = iaasClient.PartialUpdateNetwork(context.Background(), projectId, *network.NetworkId).PartialUpdateNetworkPayload(updateNetworkPayload).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when calling `PartialUpdateNetwork`: %v\n", err)
		os.Exit(1)
	}

	_, err = wait.UpdateNetworkWaitHandler(context.Background(), iaasClient, projectId, *network.NetworkId).WaitWithContext(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when waiting for update: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[IaaS API] Network has been successfully updated.\n")

	// Delete a network
	err = iaasClient.DeleteNetwork(context.Background(), projectId, *network.NetworkId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when calling `DeleteNetwork`: %v\n", err)
		os.Exit(1)
	}

	_, err = wait.DeleteNetworkWaitHandler(context.Background(), iaasClient, projectId, *network.NetworkId).WaitWithContext(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "[IaaS API] Error when waiting for deletion: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[IaaS API] Network has been successfully deleted.\n")
}
