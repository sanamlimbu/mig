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

# --- Public Route Table ---

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }

  tags = {
    Name = "${local.project_name}-public-route-table"
  }
}

resource "aws_route_table_association" "public" {
  count          = local.azs_count
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# --- EC2 Instance for API Server and NATS ---

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

resource "aws_instance" "mig" {
  ami                         = data.aws_ami.ubuntu_24_04.id
  instance_type               = var.mig_instance_type
  subnet_id                   = aws_subnet.public[0].id
  associate_public_ip_address = true
  vpc_security_group_ids      = [aws_security_group.mig_sg.id]
  tags = {
    Name = "mig-server"
  }
}

# --- Security Group for Mig EC2 Instance ---

resource "aws_security_group" "mig_sg" {
  vpc_id = aws_vpc.main.id
  name   = "${local.project_name}-instance-sg"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${local.project_name}-instance-sg"
  }
}

