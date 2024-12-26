resource "aws_ecr_repository" "repo" {
  for_each     = local.services
  name         = each.key
  force_delete = true
}
