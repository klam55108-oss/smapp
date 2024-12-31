output "ssh_key" {
  value     = tls_private_key.private_key.private_key_pem
  sensitive = true
}

output "instances" {
  value = [
    for i in range(length(var.instances)) :
    { ip = aws_instance.instance[i].public_ip, is_gateway = var.instances[i].is_gateway, is_manager = var.instances[i].is_manager }
  ]
}

output "elb_endpoint" {
  value = aws_lb.elb.dns_name
}

output "s3_endpoint" {
  value = aws_s3_bucket.bucket.bucket_regional_domain_name
}