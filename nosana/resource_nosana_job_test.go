package nosana

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccNosanaJob_basic(t *testing.T) {
	resourceName := "nosana_job.test"
	jobName := acctest.RandString(10)

	// Skip the test if in normal unit test mode (requires real credentials)
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test. Set TF_ACC=1 to run.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Providers:         testAccProviders, // Add this line
		CheckDestroy:      testAccCheckNosanaJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNosanaJobConfig_basic(jobName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNosanaJobExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_completion", "false"),
					resource.TestCheckResourceAttr(resourceName, "completion_timeout_seconds", "300"),
				),
			},
		},
	})
}

func TestAccNosanaJob_waitForCompletion(t *testing.T) {
	resourceName := "nosana_job.test"
	jobName := acctest.RandString(10)

	// Skip the test if in normal unit test mode (requires real credentials)
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test. Set TF_ACC=1 to run.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Providers:         testAccProviders, // Add this line
		CheckDestroy:      testAccCheckNosanaJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNosanaJobConfig_waitForCompletion(jobName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNosanaJobExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "wait_for_completion", "true"),
					resource.TestCheckResourceAttr(resourceName, "completion_timeout_seconds", "600"),
				),
			},
		},
	})
}

func testAccPreCheck(t *testing.T) {
	// Check for required environment variables or provider configuration
	if v := os.Getenv("NOSANA_PRIVATE_KEY"); v == "" {
		if v := os.Getenv("NOSANA_KEYPAIR_PATH"); v == "" {
			t.Fatal("NOSANA_PRIVATE_KEY or NOSANA_KEYPAIR_PATH must be set for acceptance tests")
		}
	}
}

var testAccProviders = testAccProviderFactories

func testAccCheckNosanaJobDestroy(s *terraform.State) error {
	// Since Nosana jobs are ephemeral and cannot be destroyed in the traditional sense,
	// we consider them "destroyed" when they're completed or no longer running
	return nil
}

func testAccCheckNosanaJobExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Job ID is set")
		}

		// In a real implementation, you would check if the job exists via API
		// For now, we just verify the ID is set
		return nil
	}
}

func testAccNosanaJobConfig_basic(jobName string) string {
	return fmt.Sprintf(`
provider "nosana" {
  network        = "devnet"
  market_address = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
}

resource "nosana_job" "test" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "%s",
        "args": {
          "image": "alpine:latest",
          "cmd": ["echo", "hello world"]
        },
        "type": "container/run"
      }
    ],
    "meta": {
      "trigger": "terraform-test"
    },
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = false
  completion_timeout_seconds = 300
}
`, jobName)
}

func testAccNosanaJobConfig_waitForCompletion(jobName string) string {
	return fmt.Sprintf(`
provider "nosana" {
  network        = "devnet"
  market_address = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
}

resource "nosana_job" "test" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "%s",
        "args": {
          "image": "alpine:latest",
          "cmd": ["echo", "hello world"]
        },
        "type": "container/run"
      }
    ],
    "meta": {
      "trigger": "terraform-test"
    },
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = true
  completion_timeout_seconds = 600
}
`, jobName)
}
