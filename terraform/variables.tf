variable "lambda_function_name" {
  type    = string
  default = "mass-exec"
}

variable "s3_bucket_name" {
  type = string
}

variable "s3_object_prefix" {
  type    = string
  default = "runs"
}

variable "ecr_repo_name" {
  type    = string
  default = "mass-exec"
}

variable "image_tag" {
  type = string
}