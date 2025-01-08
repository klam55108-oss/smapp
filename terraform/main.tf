terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
    tls = {
      source = "hashicorp/tls"
    }
  }
}

provider "aws" {
  region = var.region
}

locals {
  services = toset(["traefik", "user", "user-grpc", "post", "image", "image-grpc"])
}
