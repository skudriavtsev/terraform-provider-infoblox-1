package infoblox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	ibclient "github.com/infobloxopen/infoblox-go-client"
)

func testAccCheckARecordDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "resource_a_record" {
			continue
		}
		connector := meta.(ibclient.IBConnector)
		objMgr := ibclient.NewObjectManager(connector, "terraform_test", "test")
		rec, _ := objMgr.GetARecordByRef(rs.Primary.ID)
		if rec != nil {
			return fmt.Errorf("record not found")
		}

	}
	return nil
}

func testAccARecordCompare(t *testing.T, resPath string, expectedRec *ibclient.RecordA) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, found := s.RootModule().Resources[resPath]
		if !found {
			return fmt.Errorf("Not found: %s", resPath)
		}
		if res.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}
		meta := testAccProvider.Meta()
		connector := meta.(ibclient.IBConnector)
		objMgr := ibclient.NewObjectManager(connector, "terraform_test", "test")

		rec, _ := objMgr.GetARecordByRef(res.Primary.ID)
		if rec == nil {
			return fmt.Errorf("record not found")
		}

		if rec.Name != expectedRec.Name {
			return fmt.Errorf(
				"'fqdn' does not match: got '%s', expected '%s'",
				rec.Name,
				expectedRec.Name)
		}
		if rec.Ipv4Addr != expectedRec.Ipv4Addr {
			return fmt.Errorf(
				"'ipv4address' does not match: got '%s', expected '%s'",
				rec.Ipv4Addr, expectedRec.Ipv4Addr)
		}
		if rec.View != expectedRec.View {
			return fmt.Errorf(
				"'dns_view' does not match: got '%s', expected '%s'",
				rec.View, expectedRec.View)
		}
		if rec.UseTTL != expectedRec.UseTTL {
			return fmt.Errorf(
				"TTL usage does not match: got '%t', expected '%t'",
				rec.UseTTL, expectedRec.UseTTL)
		}
		if rec.UseTTL {
			if rec.TTL != expectedRec.TTL {
				return fmt.Errorf(
					"'ttl' usage does not match: got '%d', expected '%d'",
					rec.TTL, expectedRec.TTL)
			}
		}
		if rec.Comment != expectedRec.Comment {
			return fmt.Errorf(
				"'comment' does not match: got '%s', expected '%s'",
				rec.Comment, expectedRec.Comment)
		}
		return validateEAs(rec.Ea, expectedRec.Ea)
	}
}

func TestAccResourceARecord(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckARecordDestroy,
		Steps: []resource.TestStep{
			{
				// check that 'CREATE' operation correctly sets TTL = 0
				Config: `
					resource "infoblox_a_record" "foo0"{
						fqdn = "name0.a.com"
						ip_addr = "10.0.0.2"
						ttl = 0
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccARecordCompare(t, "infoblox_a_record.foo0", &ibclient.RecordA{
						Ipv4Addr: "10.0.0.2",
						Name:     "name0.a.com",
						View:     "default",
						TTL:      0,
						UseTTL:   true,
						Comment:  "",
						Ea:       nil,
					}),
				),
			},
			{
				// check that 'CREATE' operation correctly sets TTL = undef (inherited from a parent)
				Config: `
					resource "infoblox_a_record" "foo"{
						fqdn = "name1.a.com"
						ip_addr = "10.0.0.2"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccARecordCompare(t, "infoblox_a_record.foo", &ibclient.RecordA{
						Ipv4Addr: "10.0.0.2",
						Name:     "name1.a.com",
						View:     "default",
						TTL:      0,
						UseTTL:   false,
						Comment:  "",
						Ea:       nil,
					}),
				),
			},

			{
				// check that 'UPDATE' operation correctly sets TTL = 0
				Config: `
					resource "infoblox_a_record" "foo"{
						fqdn = "name1.a.com"
						ip_addr = "10.0.0.2"
						comment = "TTL=0"
						ttl = 0
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccARecordCompare(t, "infoblox_a_record.foo", &ibclient.RecordA{
						Ipv4Addr: "10.0.0.2",
						Name:     "name1.a.com",
						View:     "default",
						TTL:      0,
						UseTTL:   true,
						Comment:  "TTL=0",
						Ea:       nil,
					}),
				),
			},
			{
				// check that 'UPDATE' operation correctly sets TTL = undef
				Config: `
					resource "infoblox_a_record" "foo"{
						fqdn = "name1.a.com"
						ip_addr = "10.0.0.2"
						comment = "TTL=undef"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccARecordCompare(t, "infoblox_a_record.foo", &ibclient.RecordA{
						Ipv4Addr: "10.0.0.2",
						Name:     "name1.a.com",
						View:     "default",
						TTL:      0,
						UseTTL:   false,
						Comment:  "TTL=undef",
						Ea:       nil,
					}),
				),
			},
			{
				Config: `
					resource "infoblox_a_record" "foo2"{
						fqdn = "name2.b.com"
						ip_addr = "192.168.31.31"
						ttl = 10
						dns_view = "nondefault_view"
						comment = "test comment 1"
						extensible_attributes = jsonencode({
						  "Location" = "New York"
						  "Site" = "HQ"
						})
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccARecordCompare(t, "infoblox_a_record.foo2", &ibclient.RecordA{
						Ipv4Addr: "192.168.31.31",
						Name:     "name2.b.com",
						View:     "nondefault_view",
						TTL:      10,
						UseTTL:   true,
						Comment:  "test comment 1",
						Ea: ibclient.EA{
							"Location": "New York",
							"Site":     "HQ",
						},
					}),
				),
			},
			{
				Config: `
					resource "infoblox_a_record" "foo2"{
						fqdn = "name3.c.com"
						ip_addr = "10.10.0.1"
						ttl = 155
						dns_view = "nondefault_view"
						comment = "test comment 2"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccARecordCompare(t, "infoblox_a_record.foo2", &ibclient.RecordA{
						Ipv4Addr: "10.10.0.1",
						Name:     "name3.c.com",
						View:     "nondefault_view",
						TTL:      155,
						UseTTL:   true,
						Comment:  "test comment 2",
					}),
				),
			},
			{
				Config: `
					resource "infoblox_a_record" "foo2"{
						fqdn = "name3.c.com"
						ip_addr = "10.10.0.1"
						dns_view = "nondefault_view"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccARecordCompare(t, "infoblox_a_record.foo2", &ibclient.RecordA{
						Ipv4Addr: "10.10.0.1",
						Name:     "name3.c.com",
						View:     "nondefault_view",
						UseTTL:   false,
					}),
				),
			},
		},
	})
}

func testAccARecordIpChange(t *testing.T, prevIp *string, resPath string, mustBeEqual bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, found := s.RootModule().Resources[resPath]
		if !found {
			return fmt.Errorf("Not found: %s", resPath)
		}
		if res.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}
		meta := testAccProvider.Meta()
		connector := meta.(ibclient.IBConnector)
		objMgr := ibclient.NewObjectManager(connector, "terraform_test", "test")

		rec, _ := objMgr.GetARecordByRef(res.Primary.ID)
		if rec == nil {
			return fmt.Errorf("record not found")
		}

		if rec.Ipv4Addr != *prevIp && mustBeEqual {
			return fmt.Errorf("IP address for the A-record must not be changed")
		}

		if rec.Ipv4Addr == *prevIp && !mustBeEqual {
				return fmt.Errorf("IP address for the A-record must be different from the previous one")
		}

		*prevIp = rec.Ipv4Addr

		return nil
	}
}

func TestAccResourceARecordDynamic(t *testing.T) {

	var (
		ip1 string
		ip2 string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckARecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "infoblox_a_record" "dyn1"{
						fqdn = "dyn1.test.com"
						cidr = "10.20.30.0/24"
					}
					resource "infoblox_a_record" "dyn2_nondefault"{
						fqdn = "dyn2.test.com"
						cidr = "10.20.30.0/24"
						dns_view = "nondefault_view"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccARecordIpChange(t, &ip1, "infoblox_a_record.dyn1", false),
					testAccARecordIpChange(t, &ip2, "infoblox_a_record.dyn2_nondefault", false),
				),
			},

			// update must not change anything
			{
				Config: `
					resource "infoblox_a_record" "dyn1"{
						fqdn = "dyn1.test.com"
						cidr = "10.20.30.0/24"
					}
					resource "infoblox_a_record" "dyn2_nondefault"{
						fqdn = "dyn2.test.com"
						cidr = "10.20.30.0/24"
						dns_view = "nondefault_view"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccARecordIpChange(t, &ip1, "infoblox_a_record.dyn1", true),
					testAccARecordIpChange(t, &ip2, "infoblox_a_record.dyn2_nondefault", true),
				),
			},
		},
	})
}
