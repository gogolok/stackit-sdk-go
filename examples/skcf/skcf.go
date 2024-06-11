package main

import (
	"context"
	"fmt"
	"os"

	"github.com/stackitcloud/stackit-sdk-go/core/config"
	"github.com/stackitcloud/stackit-sdk-go/services/skcf"
	"github.com/stackitcloud/stackit-sdk-go/services/skcf/wait"
)

func main() {
	// Specify the project ID
	projectId := "PROJECT_ID"

	// Create a new API client, that uses default authentication and configuration
	skcfClient, err := skcf.NewAPIClient(
		config.WithRegion("eu01"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Creating API client: %v\n", err)
		os.Exit(1)
	}

	// The following operations assume you have a skcf project already created. If you dont, run:
	// createProjectResponse, err := skcfClient.CreateProject(context.Background(), projectId).Execute()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error when calling `CreateProject`: %v\n", err)
	// } else {
	// 	fmt.Printf("Project created with projectId: %v\n", len(*createProjectResponse.ProjectId))
	// }

	// Get the skcf clusters for your project
	getClustersResp, err := skcfClient.ListClusters(context.Background(), projectId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GetClusters`: %v\n", err)
	} else {
		fmt.Printf("Number of clusters: %v\n", len(*getClustersResp.Items))
	}

	var availableVersion string
	// Get the skcf provider options
	getOptionsResp, err := skcfClient.ListProviderOptions(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GetOptions`: %v\n", err)
	} else {
		availableVersions := *getOptionsResp.KubernetesVersions
		availableVersion = *availableVersions[0].Version
		fmt.Printf("First available version: %v\n", availableVersion)
	}

	// Create an skcf cluster
	createInstancePayload := skcf.CreateOrUpdateClusterPayload{}
	clusterName := "cl-name"
	createClusterResp, err := skcfClient.CreateOrUpdateCluster(context.Background(), projectId, clusterName).CreateOrUpdateClusterPayload(createInstancePayload).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CreateCluster`: %v\n", err)
	} else {
		fmt.Printf("Triggered creation of cluster with name \"%s\".\n", *createClusterResp.Name)
	}

	// Wait for cluster creation to complete
	_, err = wait.CreateOrUpdateClusterWaitHandler(context.Background(), skcfClient, projectId, clusterName).WaitWithContext(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CreateOrUpdateCluster`: %v\n", err)
	} else {
		fmt.Printf("Cluster created.\n")
	}
}
