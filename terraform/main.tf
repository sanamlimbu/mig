provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Name  = local.project_name
      Stage = var.stage_name
    }
  }
}

locals {
  project_name = "mig"
}

# --- VPC ---

data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  azs_count = 2
  azs_names = data.aws_availability_zones.available.names
}

resource "aws_vpc" "main" {
  cidr_block           = "10.10.0.0/22"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_subnet" "public" {
  count             = local.azs_count
  vpc_id            = aws_vpc.main.id
  availability_zone = local.azs_names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.main.cidr_block, 2, count.index)
  tags = {
    Name = "${local.project_name}-public-subnet-${local.azs_names[count.index]}"
  }
}

resource "aws_subnet" "private" {
  count             = local.azs_count
  vpc_id            = aws_vpc.main.id
  availability_zone = local.azs_names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.main.cidr_block, 2, count.index + local.azs_count)
  tags = {
    Name = "${local.project_name}-private-subnet-${local.azs_names[count.index]}"
  }
}

# --- Internet Gateway ---

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id
}

# --- Main Route Table of Mig VPC ---

resource "aws_default_route_table" "rt" {
  default_route_table_id = aws_vpc.main.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }
  tags = {
    Name = "${local.project_name}-main-route-table"
  }
}

# --- EC2 Instance for NATS ---

data "aws_ami" "ubuntu_24_04" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  owners = ["099720109477"] #canonical
}

resource "aws_instance" "nats" {
  ami                         = data.aws_ami.ubuntu_24_04.id
  instance_type               = var.nats_instance_type
  subnet_id                   = aws_subnet.public[0].id
  associate_public_ip_address = true
  vpc_security_group_ids      = [aws_security_group.nats_sg.id]
  tags = {
    Name = "nats"
  }
}

# --- Security Group for NATS Instance ---

resource "aws_security_group" "nats_sg" {
  vpc_id = aws_vpc.main.id
  name   = "${local.project_name}-nats-sg"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 4222
    to_port     = 4222
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${local.project_name}-nats-sg"
  }
}

