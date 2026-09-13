terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.4"
    }
  }

  # "state" = memória do que já foi criado (terraform.tfstate).
  # produção real, configurar um backend remoto (bucket S3) — mesmo padrão
  # já usado (comentado) nos outros projetos Terraform deste sistema.
  # backend "s3" {
  #   bucket = "autoshop-terraform-state"
  #   key    = "lambda-auth/terraform.tfstate"
  #   region = "us-east-1"
  # }
}

provider "aws" {
  region = var.aws_region
}
