package vsphere

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"regexp"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/license"
)

var (
	testAccLabels = map[string]string{
		"VpxClientLicenseLabel": "Hello World",
		"TestTitle":             "FooBar",
	}
)

func TestAccVSphereLicenseBasic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccVSpherePreLicenseBasicCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereLicenseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVSphereLicenseBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereLicenseExists("vsphere_license.foo"),
				),
			},
		},
	})

}

func TestAccVSphereLicenseInvalid(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccVSphereLicenseInvalidConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereLicenseNotExists("vsphere_license.foo"),
				),
				ExpectError: regexp.MustCompile("License file not found"),
			},
		},
	})

}

func TestAccVSphereLicenseWithLabels(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccVSpherePreLicenseBasicCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccVSphereLicenseWithLabelConfig(testAccLabels),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereLicenseWithLabelExists("vsphere_license.foo"),
				),
			},
		},
	})

}

func testAccVSphereLicenseInvalidConfig() string {

	// quite sure this key cannot be valid
	return `resource "vsphere_license" "foo" {
  					license_key = "HN422-47193-58V7M-03086-0JAN2"
			}`
}

func testAccVSphereLicenseWithLabelConfig(labels map[string]string) string {

	// precheck already checks if this is present or not
	key := os.Getenv("VSPHERE_LICENSE")

	labelString := labelToString(labels)

	return fmt.Sprintf(`resource "vsphere_license" "foo" {
							license_key = "%s"

							labels {
								%s
							}		 	
						}`,
		key, labelString)
}

func labelToString(labels map[string]string) string {
	val := ""
	for key, value := range labels {
		val += fmt.Sprintf(`
			%s = "%s"
		`, key, value)

	}
	return val
}

func testAccVSphereLicenseBasicConfig() string {

	// precheck already checks if this is present or not
	key := os.Getenv("VSPHERE_LICENSE")

	return fmt.Sprintf(`resource "vsphere_license" "foo" {
  		license_key = "%s"
		}
	`, key)

}

func testAccVSphereLicenseDestroy(s *terraform.State) error {

	client := testAccProvider.Meta().(*govmomi.Client)

	manager := license.NewManager(client.Client)

	message := ""
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_license" {
			continue
		}

		key := rs.Primary.ID
		if isKeyPresent(key, manager) {
			message += fmt.Sprintf("%s is still present on the server", key)
		}

	}
	if message != "" {
		return errors.New(message)
	}

	return nil
}

func testAccVSphereLicenseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}

		client := testAccProvider.Meta().(*govmomi.Client)
		manager := license.NewManager(client.Client)

		if !isKeyPresent(rs.Primary.ID, manager) {
			return fmt.Errorf("%s key not found on the server", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVSphereLicenseNotExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]

		if ok {
			return fmt.Errorf("%s key should not be present on the server", name)
		}

		return nil
	}
}

func testAccVSpherePreLicenseBasicCheck(t *testing.T) {
	if key := os.Getenv("VSPHERE_LICENSE"); key == "" {
		t.Fatal("VSPHERE_LICENSE must be set for acceptance test")
	}
}

func testAccVSphereLicenseWithLabelExists(name string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}

		client := testAccProvider.Meta().(*govmomi.Client)
		manager := license.NewManager(client.Client)

		if !isKeyPresent(rs.Primary.ID, manager) {
			return fmt.Errorf("%s key not found on the server", rs.Primary.ID)
		}

		info := getLicenseInfoFromKey(rs.Primary.ID, manager)

		if len(info.Labels) == 0 {
			return fmt.Errorf("The labels were not set for the key %s", info.LicenseKey)
		}

		if len(info.Labels) != 2 {
			return fmt.Errorf(`The number of labels on the server are incorrect. Expected 2 Got %d`,
				len(info.Labels))
		}

		return nil
	}

}
