package main

import (
	"context"
	"dagger/dagger-module-ci-cd/internal/dagger"
	"dagger/dagger-module-ci-cd/utils"
	"fmt"
	"path/filepath"
)

type DaggerModuleCiCd struct{}

func terraformBase(
	sshHostDir *dagger.Directory,
	deployHostDir *dagger.Directory,
	awsRegion string,
	buildVersion string,
	appName string,
	tfInitConfig string,
	tenant string,
) *dagger.Container {
	return dag.
		Container().
		From(utils.TF_IMG).
		WithMountedDirectory(utils.CONTAINER_SSH_DIR, sshHostDir).
		WithMountedDirectory(utils.WORK_DIR, deployHostDir).
		WithWorkdir(utils.WORK_DIR).
		WithEnvVariable("TF_VAR_region", awsRegion).
		WithEnvVariable("TF_VAR_build_version", buildVersion).
		WithEnvVariable("TF_VAR_service_name", appName).
		WithExec([]string{"init", tfInitConfig}).
		WithExec([]string{"workspace", "select", "-or-create", tenant}).
		WithExec([]string{"fmt", "-check"}).
		WithExec([]string{"validate"})
}

// Run terraform apply to deploy and manage AWS resources.
func (m *DaggerModuleCiCd) CdTerraformDeploy(
	ctx context.Context,
	s3BucketName string,
	appName string,
	buildVersion string,
	tenant string,
	// +optional
	// +default="us-west-2"
	awsRegion string,
) (string, error) {
	deployHostDir := dag.Directory().Directory("deploy")
	sshHostDir := dag.Directory().Directory("/home/ec2-user/.ssh")
	tfInitConfig := utils.TfBucketConfig(s3BucketName)
	tfPlanConfig := utils.TfVarFileConfig(tenant)

	return terraformBase(
		sshHostDir, deployHostDir, awsRegion, buildVersion, appName, tfInitConfig, tenant,
	).
		WithExec([]string{"plan", tfPlanConfig, "-out", "tfplan"}).
		WithExec([]string{"apply", "tfplan"}).
		Stdout(ctx)
}

// Run terraform destroy to clean service's AWS resources.
func (m *DaggerModuleCiCd) CdTerraformDestroy(
	ctx context.Context,
	s3BucketName string,
	appName string,
	buildVersion string,
	tenant string,
	// +optional
	// +default="us-west-2"
	awsRegion string,
) (string, error) {
	deployHostDir := dag.Directory().Directory("deploy")
	sshHostDir := dag.Directory().Directory("/home/ec2-user/.ssh")
	tfInitConfig := utils.TfBucketConfig(s3BucketName)
	tfPlanConfig := utils.TfVarFileConfig(tenant)

	return terraformBase(
		sshHostDir, deployHostDir, awsRegion, buildVersion, appName, tfInitConfig, tenant,
	).
		WithExec([]string{"plan", "-destroy", tfPlanConfig, "-out", "tfplan"}).
		WithExec([]string{"apply", "tfplan"}).
		Stdout(ctx)
}

// Build nodejs service.
func (m *DaggerModuleCiCd) CiNodejsBuild(
	ctx context.Context,
	githubToken string,
	// +optional
	// +default="18"
	nodeVersion string,
) (string, error) {
	src := dag.Directory().Directory(".")
	nodejsImage := utils.GetNodejsImage(nodeVersion)

	return dag.
		Container().
		WithEnvVariable("GITHUB_TOKEN", githubToken).
		From(nodejsImage).
		WithMountedDirectory(utils.WORK_DIR, src).
		WithWorkdir(utils.WORK_DIR).
		WithExec([]string{"apk", "update"}).
		WithExec([]string{"apk", "add", "--no-cache", "bash"}).
		WithExec([]string{"yarn", "install", "--frozen-lockfile"}).
		WithExec([]string{"yarn", "build"}).
		Stdout(ctx)
}

// Publish image to artifactory.
func (m *DaggerModuleCiCd) CiNodejsPublishImage(
	ctx context.Context,
	githubToken string,
	ecrImgName string,
	gitCommitId string,
) (string, error) {
	imgPath := fmt.Sprintf("%s:default-%s", ecrImgName, gitCommitId)
	buildOps := dagger.DirectoryDockerBuildOpts{
		Dockerfile: "./Dockerfile.prod",
		BuildArgs: []dagger.BuildArg{
			{Name: "GITHUB_TOKEN", Value: githubToken},
		},
	}

	return dag.
		Directory().
		Directory(".").
		DockerBuild(buildOps).
		Publish(ctx, imgPath)
}

// Deploy service infra.
func (m *DaggerModuleCiCd) CiServiceInfra(
	ctx context.Context,
	bucketName string,
	appName string,
	env string,
) (string, error) {
	src := dag.CurrentModule().Source().Directory(".")
	tfWorkDir := filepath.Join(utils.MNT_PREFIX, "terraform-svc-shared-infra")
	s3KeyBackend := fmt.Sprintf("-backend-config=key=services/shared/%s", appName)
	tfInitConfig := fmt.Sprintf("-backend-config=bucket=%s", bucketName)

	return dag.
		Container().
		WithEnvVariable("TF_VAR_service_name", appName).
		From(utils.TF_IMG).
		WithMountedDirectory(utils.MNT_PREFIX, src).
		WithWorkdir(tfWorkDir).
		WithExec([]string{"init", tfInitConfig, s3KeyBackend}).
		WithExec([]string{"workspace", "select", "-or-create", env}).
		WithExec([]string{"fmt", "-check"}).
		WithExec([]string{"validate"}).
		WithExec([]string{"plan", "-out", "tfplan"}).
		WithExec([]string{"apply", "tfplan"}).
		Stdout(ctx)
}
