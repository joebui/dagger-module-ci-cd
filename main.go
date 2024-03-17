package main

import (
	"dagger/dagger-module-ci-cd/internal/dagger"
	"dagger/dagger-module-ci-cd/utils"
)

type DaggerModuleCiCd struct {
	container *dagger.Container
}

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

// Refer to ci.go, cd.go for dagger functions.
