resource "aws_default_vpc" "default" {
  tags = {
    Name = "Default VPC"
  }
}

resource "aws_default_subnet" "default" {
  for_each          = var.elb_availability_zones
  availability_zone = each.key

  tags = {
    Name = "Default subnet in ${each.key}"
  }
}
