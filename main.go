// A generated module for DaggerModuleCiCd functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/dagger-module-ci-cd/internal/dagger"
	"dagger/dagger-module-ci-cd/utils"
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
func (m *DaggerModuleCiCd) TerraformDeploy(
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
func (m *DaggerModuleCiCd) TerraformDestroy(
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
