variable "region" {
  type = string
}

# Availability zones to place load balancers in
variable "elb_availability_zones" {
  type = set(string)
}

variable "instances" {
  type = list(object({
    name              = string
    availability_zone = string
    # Whether the node gets HTTP traffic from ELB
    is_gateway = bool
    # Whether the node is a Swarm manager
    is_manager = bool
  }))
}

variable "bucket_name" {
  type = string
}
