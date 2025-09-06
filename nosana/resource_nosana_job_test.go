package nosana

import (
	"context"
	"fmt"
	"os"
	"strings"
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
		t.Skip("TF_ACC not set, skipping acceptance test.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
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
		t.Skip("TF_ACC not set, skipping acceptance test.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
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

func TestAccNosanaJob_update(t *testing.T) {
	resourceName := "nosana_job.test"
	jobName := acctest.RandString(10)

	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckNosanaJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNosanaJobConfig_basic(jobName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNosanaJobExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replicas", "1"),
					resource.TestCheckResourceAttr(resourceName, "timeout", "300"),
				),
			},
			{
				Config: testAccNosanaJobConfig_update(jobName, 2, 600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNosanaJobExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replicas", "2"),
					resource.TestCheckResourceAttr(resourceName, "timeout", "600"),
				),
			},
		},
	})
}

func testAccPreCheck(t *testing.T) {
	// Check for required environment variables or provider configuration
	if v := os.Getenv("NOSANA_MARKET_ADDRESS"); v == "" {
		t.Fatal("NOSANA_MARKET_ADDRESS must be set for acceptance tests")
	}
}

func testAccCheckNosanaJobDestroy(s *terraform.State) error {
	client := testAccProvider().Meta().(*nosanaClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "nosana_job" {
			continue
		}

		// Try to get the deployment to see if it still exists
		_, err := client.APIClient.GetDeployment(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Nosana job %s still exists", rs.Primary.ID)
		}

		// Check if it's a 404 error (expected for destroyed resources)
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("Unexpected error checking for destroyed job: %s", err)
		}
	}

	return nil
}

func testAccCheckNosanaJobExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Nosana job ID is set")
		}

		client := testAccProvider().Meta().(*nosanaClient)
		_, err := client.APIClient.GetDeployment(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error fetching Nosana job: %s", err)
		}

		return nil
	}
}

func testAccNosanaJobConfig_basic(jobName string) string {
	return fmt.Sprintf(`
provider "nosana" {
  use_local_wallet = true
  market_address   = "%s"
}

resource "nosana_job" "test" {
  job_content = jsonencode({
    "ops": [
      {
        "id": "test-job",
        "args": {
          "env": {
            "TEST_VAR": "test_value"
          },
          "image": "alpine:latest",
          "expose": 8080
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
  replicas = 1
  timeout  = 300
  strategy = "SIMPLE"

  wait_for_completion        = false
  completion_timeout_seconds = 300
}
`, os.Getenv("NOSANA_MARKET_ADDRESS"))
}

func testAccNosanaJobConfig_waitForCompletion(jobName string) string {
	return fmt.Sprintf(`
provider "nosana" {
  use_local_wallet = true
  market_address   = "%s"
}

resource "nosana_job" "test" {
  job_content = jsonencode({
    "ops": [
      {
        "id": "test-job-wait",
        "args": {
          "env": {
            "TEST_VAR": "test_value"
          },
          "image": "alpine:latest",
          "expose": 8080
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
  replicas = 1
  timeout  = 600
  strategy = "SIMPLE"

  wait_for_completion        = true
  completion_timeout_seconds = 600
}
`, os.Getenv("NOSANA_MARKET_ADDRESS"))
}

func testAccNosanaJobConfig_update(jobName string, replicas, timeout int) string {
	return fmt.Sprintf(`
provider "nosana" {
  use_local_wallet = true
  market_address   = "%s"
}

resource "nosana_job" "test" {
  job_content = jsonencode({
    "ops": [
      {
        "id": "test-job-update",
        "args": {
          "env": {
            "TEST_VAR": "test_value"
          },
          "image": "alpine:latest",
          "expose": 8080
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
  replicas = %d
  timeout  = %d
  strategy = "SIMPLE"

  wait_for_completion        = false
  completion_timeout_seconds = 300
}
`, os.Getenv("NOSANA_MARKET_ADDRESS"), replicas, timeout)
}
