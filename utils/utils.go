package utils

import "fmt"

const (
	WORK_DIR          = "/mnt"
	TF_IMG            = "hashicorp/terraform:1.5"
	CONTAINER_SSH_DIR = "/root/.ssh"
)

func TfBucketConfig(bucketName string) string {
	return fmt.Sprintf("-backend-config=bucket=%s", bucketName)
}

func TfVarFileConfig(tenant string) string {
	return fmt.Sprintf("-var-file=%s.tfvars", tenant)
}

func GetNodejsImage(version string) string {
	return fmt.Sprintf("node:%s-alpine", version)
}
