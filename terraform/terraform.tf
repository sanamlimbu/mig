terraform {
  cloud {
    organization = "sanam-default-org"
    workspaces {
      name = "mig-prod"
    }
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.84.0"
    }
  }

  required_version = "~> 1.10.2"
}

