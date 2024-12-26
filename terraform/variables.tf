variable "region" {
  type = string
}

variable "availability_zones" {
  type = set(string)
}

variable "instance_count" {
  type = number
}

variable "bucket_name" {
  type = string
}
