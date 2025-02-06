variable "aws_region" {
  description = "AWS region for all resources."
  type        = string
  default     = "ap-southeast-2"
}

variable "stage_name" {
  description = "Deployment stage name."
  type        = string
  default     = "prod"
}

variable "db_host" {
  description = "Database host."
  type        = string
}

variable "db_user" {
  description = "Database user."
  type        = string
}

variable "db_password" {
  description = "Database password."
  type        = string
}

variable "db_name" {
  description = "Database name."
  type        = string
}

variable "db_port" {
  description = "Database port."
  type        = number
}

variable "mig_instance_type" {
  description = "EC2 instance type for api server."
  type        = string
  default     = "t2.micro"
}