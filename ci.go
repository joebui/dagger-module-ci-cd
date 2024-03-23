package main

import (
	"context"
	"dagger/dagger-module-ci-cd/internal/dagger"
	"dagger/dagger-module-ci-cd/utils"
	"fmt"
)

// Build nodejs service.
func (m *DaggerModuleCiCd) CiNodejsBuild(
	ctx context.Context,
	githubToken string,
	// +optional
	// +default="18"
	nodeVersion string,
	src *dagger.Directory,
) *dagger.Directory {
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
		Directory("./dist")
}

// Deploy shared service infra.
func (m *DaggerModuleCiCd) CiServiceInfra(
	ctx context.Context,
	bucketName string,
	appName string,
	env string,
	src *dagger.Directory,
) (string, error) {
	s3KeyBackend := fmt.Sprintf("-backend-config=key=services/shared/%s", appName)
	tfInitConfig := fmt.Sprintf("-backend-config=bucket=%s", bucketName)

	return dag.
		Container().
		WithEnvVariable("TF_VAR_service_name", appName).
		From(utils.TF_IMG).
		WithMountedDirectory(utils.WORK_DIR, src).
		WithWorkdir(utils.WORK_DIR).
		WithExec([]string{"init", tfInitConfig, s3KeyBackend}).
		WithExec([]string{"workspace", "select", "-or-create", env}).
		WithExec([]string{"plan", "-out", "tfplan"}).
		WithExec([]string{"apply", "tfplan"}).
		Stdout(ctx)
}

// Build Next.js service.
func (m *DaggerModuleCiCd) CiNextjsBuild(ctx context.Context,
	githubToken string,
	src *dagger.Directory,
	// +optional
	// +default="18"
	nodeVersion string,
) *dagger.Directory {
	nodejsImage := utils.GetNodejsImage(nodeVersion)

	return dag.
		Container().
		WithEnvVariable("GITHUB_TOKEN", githubToken).
		From(nodejsImage).
		WithMountedDirectory(utils.WORK_DIR, src).
		WithWorkdir(utils.WORK_DIR).
		WithExec([]string{"npm", "install", "-g", "dotenv-cli"}).
		WithExec([]string{"yarn", "install", "--frozen-lockfile"}).
		WithExec([]string{"dotenv", "yarn", "build"}).
		Directory("./.next")
}

// Build SPA app.
func (m *DaggerModuleCiCd) CiSpaBuild(ctx context.Context,
	githubToken string,
	src *dagger.Directory,
	// +optional
	// +default="16"
	nodeVersion string,
) *dagger.Directory {
	nodejsImage := utils.GetNodejsImage(nodeVersion)

	return dag.
		Container().
		WithEnvVariable("GITHUB_TOKEN", githubToken).
		From(nodejsImage).
		WithMountedDirectory(utils.WORK_DIR, src).
		WithWorkdir(utils.WORK_DIR).
		WithExec([]string{"npm", "install", "-g", "dotenv-cli"}).
		WithExec([]string{"dotenv", "yarn", "install", "--frozen-lockfile"}).
		WithExec([]string{"dotenv", "yarn", "build"}).
		Directory("./webapp-static")
}
