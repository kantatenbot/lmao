output "repo_url" {
  value       = aws_ecr_repository.repo.repository_url
  description = "ECR repo URL (for docker tag)"
}