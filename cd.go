package main

import (
	"context"
	"dagger/dagger-module-ci-cd/internal/dagger"
	"dagger/dagger-module-ci-cd/utils"
)

// Run terraform apply to deploy and manage AWS resources.
func (m *DaggerModuleCiCd) CdTerraformApply(
	ctx context.Context,
	bucketName string,
	appName string,
	buildVersion string,
	tenant string,
	src *dagger.Directory,
	sshHostDir *dagger.Directory,
	// +optional
	// +default="us-west-2"
	awsRegion string,
) (string, error) {
	tfInitConfig := utils.TfBucketConfig(bucketName)
	tfPlanConfig := utils.TfVarFileConfig(tenant)

	return terraformBase(
		sshHostDir, src, awsRegion, buildVersion, appName, tfInitConfig, tenant,
	).
		WithExec([]string{"plan", tfPlanConfig, "-out", "tfplan"}).
		WithExec([]string{"apply", "tfplan"}).
		Stdout(ctx)
}

// Run terraform destroy to clean service's AWS resources.
func (m *DaggerModuleCiCd) CdTerraformDestroy(
	ctx context.Context,
	bucketName string,
	appName string,
	buildVersion string,
	tenant string,
	src *dagger.Directory,
	sshHostDir *dagger.Directory,
	// +optional
	// +default="us-west-2"
	awsRegion string,
) (string, error) {
	tfInitConfig := utils.TfBucketConfig(bucketName)
	tfPlanConfig := utils.TfVarFileConfig(tenant)

	return terraformBase(
		sshHostDir, src, awsRegion, buildVersion, appName, tfInitConfig, tenant,
	).
		WithExec([]string{"plan", "-destroy", tfPlanConfig, "-out", "tfplan"}).
		WithExec([]string{"apply", "tfplan"}).
		Stdout(ctx)
}
