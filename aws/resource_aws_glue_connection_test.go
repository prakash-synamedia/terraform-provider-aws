package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/glue/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_glue_connection", &resource.Sweeper{
		Name: "aws_glue_connection",
		F:    testSweepGlueConnections,
	})
}

func testSweepGlueConnections(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).glueconn
	catalogID := client.(*AWSClient).accountid

	input := &glue.GetConnectionsInput{
		CatalogId: aws.String(catalogID),
	}
	err = conn.GetConnectionsPages(input, func(page *glue.GetConnectionsOutput, lastPage bool) bool {
		if len(page.ConnectionList) == 0 {
			log.Printf("[INFO] No Glue Connections to sweep")
			return false
		}
		for _, connection := range page.ConnectionList {
			id := fmt.Sprintf("%s:%s", catalogID, aws.StringValue(connection.Name))

			log.Printf("[INFO] Deleting Glue Connection: %s", id)
			r := resourceAwsGlueConnection()
			d := r.Data(nil)
			d.SetId(id)

			err := r.Delete(d, client)
			if err != nil {
				log.Printf("[ERROR] Failed to delete Glue Connection %s: %s", id, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Connection sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Connections: %s", err)
	}

	return nil
}

func TestAccAWSGlueConnection_basic(t *testing.T) {
	var connection glue.Connection

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_connection.test"

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSGlueConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueConnectionConfig_Required(rName, jdbcConnectionUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("connection/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.JDBC_CONNECTION_URL", jdbcConnectionUrl),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.PASSWORD", "testpassword"),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.USERNAME", "testusername"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "physical_connection_requirements.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSGlueConnection_tags(t *testing.T) {
	var connection glue.Connection

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_connection.test"

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSGlueConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueConnectionConfigTags1(rName, jdbcConnectionUrl, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSGlueConnectionConfigTags2(rName, jdbcConnectionUrl, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGlueConnectionConfigTags1(rName, jdbcConnectionUrl, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGlueConnection_MongoDB(t *testing.T) {
	var connection glue.Connection

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_connection.test"

	connectionUrl := fmt.Sprintf("mongodb://%s:27017/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSGlueConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueConnectionConfig_MongoDB(rName, connectionUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.CONNECTION_URL", connectionUrl),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.USERNAME", "testusername"),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.PASSWORD", "testpassword"),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "MONGODB"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "physical_connection_requirements.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSGlueConnection_Kafka(t *testing.T) {
	var connection glue.Connection

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_connection.test"

	bootstrapServers := fmt.Sprintf("%s:9094,%s:9094", acctest.RandomDomainName(), acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSGlueConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueConnectionConfig_Kafka(rName, bootstrapServers),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.KAFKA_BOOTSTRAP_SERVERS", bootstrapServers),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "KAFKA"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "physical_connection_requirements.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSGlueConnection_Network(t *testing.T) {
	var connection glue.Connection

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSGlueConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueConnectionConfig_Network(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "NETWORK"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "physical_connection_requirements.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "physical_connection_requirements.0.availability_zone"),
					resource.TestCheckResourceAttr(resourceName, "physical_connection_requirements.0.security_group_id_list.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "physical_connection_requirements.0.subnet_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSGlueConnection_Description(t *testing.T) {
	var connection glue.Connection

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_connection.test"

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSGlueConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueConnectionConfig_Description(rName, jdbcConnectionUrl, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "description", "First Description"),
				),
			},
			{
				Config: testAccAWSGlueConnectionConfig_Description(rName, jdbcConnectionUrl, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "description", "Second Description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSGlueConnection_MatchCriteria(t *testing.T) {
	var connection glue.Connection

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_connection.test"

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSGlueConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueConnectionConfig_MatchCriteria_First(rName, jdbcConnectionUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.0", "criteria1"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.1", "criteria2"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.2", "criteria3"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.3", "criteria4"),
				),
			},
			{
				Config: testAccAWSGlueConnectionConfig_MatchCriteria_Second(rName, jdbcConnectionUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.0", "criteria1"),
				),
			},
			{
				Config: testAccAWSGlueConnectionConfig_MatchCriteria_Third(rName, jdbcConnectionUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.0", "criteria2"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.1", "criteria3"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.2", "criteria4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSGlueConnection_PhysicalConnectionRequirements(t *testing.T) {
	var connection glue.Connection

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSGlueConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueConnectionConfig_PhysicalConnectionRequirements(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "connection_properties.%", "3"),
					resource.TestCheckResourceAttrSet(resourceName, "connection_properties.JDBC_CONNECTION_URL"),
					resource.TestCheckResourceAttrSet(resourceName, "connection_properties.PASSWORD"),
					resource.TestCheckResourceAttrSet(resourceName, "connection_properties.USERNAME"),
					resource.TestCheckResourceAttr(resourceName, "match_criteria.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "physical_connection_requirements.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "physical_connection_requirements.0.availability_zone"),
					resource.TestCheckResourceAttr(resourceName, "physical_connection_requirements.0.security_group_id_list.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "physical_connection_requirements.0.subnet_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSGlueConnection_disappears(t *testing.T) {
	var connection glue.Connection

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_connection.test"

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSGlueConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueConnectionConfig_Required(rName, jdbcConnectionUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueConnectionExists(resourceName, &connection),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsGlueConnection(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsGlueConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSGlueConnectionExists(resourceName string, connection *glue.Connection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Connection ID is set")
		}

		conn := acctest.Provider.Meta().(*AWSClient).glueconn
		catalogID, connectionName, err := decodeGlueConnectionID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := finder.ConnectionByName(conn, connectionName, catalogID)

		if err != nil {
			return err
		}

		*connection = *output

		return nil
	}
}

func testAccCheckAWSGlueConnectionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_connection" {
			continue
		}

		conn := acctest.Provider.Meta().(*AWSClient).glueconn
		catalogID, connectionName, err := decodeGlueConnectionID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = finder.ConnectionByName(conn, connectionName, catalogID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Glue Connection %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSGlueConnectionConfig_Description(rName, jdbcConnectionUrl, description string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name        = %[1]q
  description = %[2]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[3]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }

}
`, rName, description, jdbcConnectionUrl)
}

func testAccAWSGlueConnectionConfig_MatchCriteria_First(rName, jdbcConnectionUrl string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name = %[1]q

  match_criteria = ["criteria1", "criteria2", "criteria3", "criteria4"]

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}
`, rName, jdbcConnectionUrl)
}

func testAccAWSGlueConnectionConfig_MatchCriteria_Second(rName, jdbcConnectionUrl string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name = %[1]q

  match_criteria = ["criteria1"]

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}
`, rName, jdbcConnectionUrl)
}

func testAccAWSGlueConnectionConfig_MatchCriteria_Third(rName, jdbcConnectionUrl string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name = "%s"

  match_criteria = ["criteria2", "criteria3", "criteria4"]

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}
`, rName, jdbcConnectionUrl)
}

func testAccAWSGlueConnectionConfig_PhysicalConnectionRequirements(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-glue-connection-base"
  }
}

resource "aws_security_group" "test" {
  name   = "%[1]s"
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 1
    protocol  = "tcp"
    self      = true
    to_port   = 65535
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-glue-connection-base"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "%[1]s"
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = "%[1]s"
  database_name                   = "gluedatabase"
  db_cluster_parameter_group_name = "default.aurora-mysql5.7"
  db_subnet_group_name            = aws_db_subnet_group.test.name
  engine                          = "aurora-mysql"
  master_password                 = "gluepassword"
  master_username                 = "glueusername"
  skip_final_snapshot             = true
  vpc_security_group_ids          = [aws_security_group.test.id]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  engine             = "aurora-mysql"
  identifier         = "%[1]s"
  instance_class     = "db.t2.medium"
}

resource "aws_glue_connection" "test" {
  connection_properties = {
    JDBC_CONNECTION_URL = "jdbc:mysql://${aws_rds_cluster.test.endpoint}/${aws_rds_cluster.test.database_name}"
    PASSWORD            = aws_rds_cluster.test.master_password
    USERNAME            = aws_rds_cluster.test.master_username
  }

  name = "%[1]s"

  physical_connection_requirements {
    availability_zone      = aws_subnet.test[0].availability_zone
    security_group_id_list = [aws_security_group.test.id]
    subnet_id              = aws_subnet.test[0].id
  }
}
`, rName)
}

func testAccAWSGlueConnectionConfig_Required(rName, jdbcConnectionUrl string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}
`, rName, jdbcConnectionUrl)
}

func testAccAWSGlueConnectionConfigTags1(rName, jdbcConnectionUrl, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, jdbcConnectionUrl, tagKey1, tagValue1)
}

func testAccAWSGlueConnectionConfigTags2(rName, jdbcConnectionUrl, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, jdbcConnectionUrl, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSGlueConnectionConfig_MongoDB(rName, connectionUrl string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_type = "MONGODB"
  connection_properties = {
    CONNECTION_URL = %[2]q
    PASSWORD       = "testpassword"
    USERNAME       = "testusername"
  }
}
`, rName, connectionUrl)
}

func testAccAWSGlueConnectionConfig_Kafka(rName, bootstrapServers string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_type = "KAFKA"
  connection_properties = {
    KAFKA_BOOTSTRAP_SERVERS = %[2]q
  }
}
`, rName, bootstrapServers)
}

func testAccAWSGlueConnectionConfig_Network(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-glue-connection-network"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-glue-connection-network"
  }
}

resource "aws_security_group" "test" {
  name   = "%[1]s"
  vpc_id = aws_vpc.test.id

  ingress {
    protocol  = "tcp"
    self      = true
    from_port = 1
    to_port   = 65535
  }
}

resource "aws_glue_connection" "test" {
  connection_type = "NETWORK"
  name            = "%[1]s"

  physical_connection_requirements {
    availability_zone      = aws_subnet.test.availability_zone
    security_group_id_list = [aws_security_group.test.id]
    subnet_id              = aws_subnet.test.id
  }
}
`, rName)
}
