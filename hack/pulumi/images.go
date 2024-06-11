package main

import (
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func buildAndPublishImages(
	ctx *pulumi.Context,
	crUrl string,
	token pulumi.StringOutput,
	images map[string]pulumi.Output,
	dexUrl string,
	enduroUrl string,
	oidcClientId pulumi.StringOutput,
) error {
	// Setup DigitalOcean container registry URL and credentials.
	registry := token.ApplyT(func(token string) docker.Registry {
		return docker.Registry{
			Server:   &crUrl,
			Username: &token,
			Password: &token,
		}
	}).(docker.RegistryOutput)

	// Build and publish enduro image.
	enduroImage, err := docker.NewImage(ctx, "enduro-image",
		&docker.ImageArgs{
			Build: &docker.DockerBuildArgs{
				Context: pulumi.String("../.."),
			},
			ImageName: pulumi.String(crUrl + "/artefactual/enduro"),
			Registry:  registry,
		},
	)
	if err != nil {
		return err
	}

	// Build and publish enduro-a3m-worker image.
	enduroA3mWorkerImage, err := docker.NewImage(ctx, "enduro-a3m-worker-image",
		&docker.ImageArgs{
			Build: &docker.DockerBuildArgs{
				Context: pulumi.String("../.."),
				Target:  pulumi.String("enduro-a3m-worker"),
			},
			ImageName: pulumi.String(crUrl + "/artefactual/enduro-a3m-worker"),
			Registry:  registry,
		},
	)
	if err != nil {
		return err
	}

	// Build and publish enduro-dashboard image.
	enduroDashboardImage, err := docker.NewImage(ctx, "enduro-dashboard-image",
		&docker.ImageArgs{
			Build: &docker.DockerBuildArgs{
				Context: pulumi.String("../../dashboard"),
				Args: pulumi.StringMap{
					"oidcProviderUrl": pulumi.String(dexUrl),
					"oidcRedirectUrl": pulumi.String(enduroUrl + "/user/signin-callback"),
					"oidcClientId":    oidcClientId,
				},
			},
			ImageName: pulumi.String(crUrl + "/artefactual/enduro-dashboard"),
			Registry:  registry,
		},
	)
	if err != nil {
		return err
	}

	// Update the images map with the built image names.
	images["enduro"] = enduroImage.ImageName
	images["enduro-a3m-worker"] = enduroA3mWorkerImage.ImageName
	images["enduro-dashboard"] = enduroDashboardImage.ImageName

	return nil
}
