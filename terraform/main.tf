module "lambda" {
  source = "terraform-aws-modules/lambda/aws"

  function_name = var.lambda_function_name
  description   = "mass-exec executor"

  timeout     = 900
  memory_size = 1024

  create_package = false

  image_uri    = module.image.image_uri
  package_type = "Image"

  environment_variables = {
    MASS_EXEC_BUCKET_NAME   = var.s3_bucket_name
    MASS_EXEC_OBJECT_PREFIX = var.s3_object_prefix
  }
}

module "image" {
  source = "terraform-aws-modules/lambda/aws//modules/docker-build"
  create_ecr_repo = false
  ecr_repo        = aws_ecr_repository.repo.name
  image_tag       = var.image_tag
  source_path     = "../" # this works because terraform is in the .dockerignore
  platform        = "linux/amd64"
}

resource "aws_ecr_repository" "repo" {
  name = var.ecr_repo_name
}

resource "aws_s3_bucket" "bucket" {
  bucket = var.s3_bucket_name
  acl    = "private"

  lifecycle_rule {
    id      = "expiry"
    enabled = true

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "GLACIER"
    }
  }

}

// give lambda permission to upload output to s3
resource "aws_iam_role_policy" "lambda_role_policy" {
  name   = "${var.lambda_function_name}-lambda-role-policy"
  role   = module.lambda.lambda_role_name
  policy = data.aws_iam_policy_document.lambda_role_policy.json
}

data "aws_iam_policy_document" "lambda_role_policy" {
  statement {
    sid    = "AllowBucketAccess"
    effect = "Allow"
    actions = [
      "s3:PutObject",
    ]
    resources = ["arn:aws:s3:::${var.s3_bucket_name}/*"]
  }
}