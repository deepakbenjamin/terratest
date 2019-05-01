package test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"terratest/modules/aws"
	"terratest/modules/random"
	"terratest/modules/terraform"
	"testing"
)

// An example of how to test the Terraform module in examples/terraform-aws-ecs-example using Terratest.
func TestTerraformAwsEcsExample(t *testing.T) {
	t.Parallel()

	expectedClusterName := fmt.Sprintf("terratest-aws-ecs-example-cluster-%s", random.UniqueId())
	expectedServiceName := fmt.Sprintf("terratest-aws-ecs-example-service-%s", random.UniqueId())

	// Pick a random AWS region to test in. This helps ensure your code works in all regions.
	awsRegion := aws.GetRandomStableRegion(t, []string{"us-east-1", "eu-west-1"}, nil)

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../examples/terraform-aws-ecs-example",

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"cluster_name": expectedClusterName,
			"service_name": expectedServiceName,
		},

		// Environment variables to set when running Terraform
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": awsRegion,
		},
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the value of an output variable
	taskDefinition := terraform.Output(t, terraformOptions, "task_definition")

	// Look up the ECS cluster by name
	cluster := aws.GetEcsCluster(t, awsRegion, expectedClusterName)

	assert.Equal(t, int64(1), *cluster.ActiveServicesCount)

	// Look up the ECS service by name
	service := aws.GetEcsService(t, awsRegion, expectedClusterName, expectedServiceName)

	assert.Equal(t, int64(0), *service.DesiredCount)
	assert.Equal(t, "FARGATE", *service.LaunchType)

	// Look up the ECS task definition by ARN
	task := aws.GetEcsTaskDefinition(t, awsRegion, taskDefinition)

	assert.Equal(t, "256", *task.Cpu)
	assert.Equal(t, "512", *task.Memory)
	assert.Equal(t, "awsvpc", *task.NetworkMode)
}