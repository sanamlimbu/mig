# Referenced from 
# https://medium.com/@vladkens/aws-ecs-cluster-on-ec2-with-terraform-2023-fdb9f6b7db07

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
      version = "~> 5.63.0"
    }
  }

  required_version = "~> 1.9.4"
}

locals {
  project_name = "mig"
}

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

variable "db_password" {
  description = "Database password."
  type        = string
}

variable "db_username" {
  description = "Database username."
  type        = string
  default     = "postgres"
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Name  = local.project_name
      Stage = var.stage_name
    }
  }
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
  cidr_block           = "10.10.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = {
    Name = "${local.project_name}-vpc"
  }
}

resource "aws_subnet" "public" {
  count             = local.azs_count
  vpc_id            = aws_vpc.main.id
  availability_zone = local.azs_names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.main.cidr_block, 8, 10 + count.index)
  tags = {
    Name = "${local.project_name}-public-subnet-${local.azs_names[count.index]}"
  }
}

resource "aws_subnet" "private" {
  count             = local.azs_count
  vpc_id            = aws_vpc.main.id
  availability_zone = local.azs_names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.main.cidr_block, 8, 20 + count.index)
  tags = {
    Name = "${local.project_name}-private-subnet-${local.azs_names[count.index]}"
  }
}

# --- Internet Gateway ---

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${local.project_name}-internet-gateway"
  }
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

# --- PostgreSQL Instance

resource "aws_db_instance" "postgres" {
  allocated_storage    = 20
  db_subnet_group_name = aws_db_subnet_group.postgres.name
  engine               = "postgres"
  engine_version       = "16.3"
  identifier           = "${local.project_name}-postgres-db"
  instance_class       = "db.t4g.micro"
  password             = var.db_password
  skip_final_snapshot  = true
  storage_encrypted    = false
  publicly_accessible  = false
  username             = var.db_username
  apply_immediately    = true
}

resource "aws_db_subnet_group" "postgres" {
  name_prefix = local.project_name
  subnet_ids  = [for s in aws_subnet.private : s.id]
}

output "db_arn" {
  value = aws_db_instance.postgres.arn
}

# --- ECS Cluster ---

resource "aws_ecs_cluster" "mig" {
  name = "${local.project_name}-ecs-cluster"
  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# --- ECS Node Role ---

data "aws_iam_policy_document" "ecs_node_doc" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "ecs_node_role" {
  name_prefix        = "${local.project_name}-ecs-node-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_node_doc.json
}

resource "aws_iam_role_policy_attachment" "ecs_node_role_policy" {
  role       = aws_iam_role.ecs_node_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_node" {
  name_prefix = "${local.project_name}-ecs-node-profile"
  path        = "/ecs/instance/"
  role        = aws_iam_role.ecs_node_role.name
}

# --- ECS Node SG ---

resource "aws_security_group" "ecs_node_sg" {
  name_prefix = "${local.project_name}-ecs-node-sg-"
  vpc_id      = aws_vpc.main.id
}

resource "aws_vpc_security_group_egress_rule" "ecs_node_allow_all_traffic_ipv4" {
  security_group_id = aws_security_group.ecs_node_sg.id
  from_port         = 0
  to_port           = 65535
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
}



# --- ECS Launch Template ---

data "aws_ssm_parameter" "ecs_node_ami" {
  name = "/aws/service/ecs/optimized-ami/amazon-linux-2/recommended/image_id"
}

resource "aws_launch_template" "ecs_ec2" {
  name_prefix            = "${local.project_name}-ecs-ec2-"
  image_id               = data.aws_ssm_parameter.ecs_node_ami.value
  instance_type          = "t2.micro"
  vpc_security_group_ids = [aws_security_group.ecs_node_sg.id]

  iam_instance_profile {
    arn = aws_iam_instance_profile.ecs_node.arn
  }

  monitoring {
    enabled = true
  }

  user_data = base64encode(<<-EOF
      #!/bin/bash
      echo ECS_CLUSTER=${aws_ecs_cluster.mig.name} >> /etc/ecs/ecs.config;
    EOF
  )
}

# --- ECS ASG ---

resource "aws_autoscaling_group" "ecs" {
  name_prefix               = "${local.project_name}-ecs-asg-"
  vpc_zone_identifier       = [for s in aws_subnet.public : s.id]
  min_size                  = 1
  max_size                  = 2
  health_check_grace_period = 0
  health_check_type         = "EC2"
  desired_capacity          = 1
  protect_from_scale_in     = false

  launch_template {
    id      = aws_launch_template.ecs_ec2.id
    version = "$Latest"
  }

  tag {
    key                 = "Name"
    value               = "${local.project_name}-ecs-cluster"
    propagate_at_launch = true
  }

  tag {
    key                 = "AmazonECSManaged"
    value               = ""
    propagate_at_launch = true
  }
}

# --- ECS Capacity Provider ---

resource "aws_ecs_capacity_provider" "main" {
  name = "${local.project_name}-ecs-ec2"

  auto_scaling_group_provider {
    auto_scaling_group_arn         = aws_autoscaling_group.ecs.arn
    managed_termination_protection = "DISABLED"

    managed_scaling {
      maximum_scaling_step_size = 2
      minimum_scaling_step_size = 1
      status                    = "ENABLED"
      target_capacity           = 100
    }
  }
}

resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name       = aws_ecs_cluster.mig.name
  capacity_providers = [aws_ecs_capacity_provider.main.name]

  default_capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.main.name
    base              = 1
    weight            = 100
  }
}

# --- ECR ---
resource "aws_ecr_repository" "mig" {
  name                 = "mig"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }
}

output "mig_repo_url" {
  value = aws_ecr_repository.mig.repository_url
}

# --- ECS Task Role ---

data "aws_iam_policy_document" "ecs_task_doc" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "ecs_task_role" {
  name_prefix        = "${local.project_name}-ecs-task-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_doc.json
}

resource "aws_iam_role" "ecs_exec_role" {
  name_prefix        = "${local.project_name}-ecs-exec-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_doc.json
}

resource "aws_iam_role_policy_attachment" "ecs_exec_role_policy" {
  role       = aws_iam_role.ecs_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# --- Cloud Watch Logs ---

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/${local.project_name}"
  retention_in_days = 14
}

# --- ALB ---

resource "aws_security_group" "http" {
  name_prefix = "${local.project_name}-alb-sg-"
  description = "Allow all HTTP/HTTPS traffic from public"
  vpc_id      = aws_vpc.main.id

  dynamic "ingress" {
    for_each = [80, 443]
    content {
      protocol    = "tcp"
      from_port   = ingress.value
      to_port     = ingress.value
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_lb" "mig" {
  name               = "mig-alb"
  load_balancer_type = "application"
  security_groups    = [aws_security_group.http.id]
  subnets            = [for s in aws_subnet.public : s.id]
}

resource "aws_lb_target_group" "mig" {
  name_prefix = "mig"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"
  port        = 80
  protocol    = "HTTP"

  health_check {
    enabled             = true
    path                = "/"
    port                = 80
    matcher             = 200
    interval            = 10
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
  }
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.mig.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.mig.arn
  }
}

output "alb_url" {
  value = aws_lb.mig.dns_name
}

# --- Task Definition ---

resource "aws_ecs_task_definition" "kafka" {
  family             = local.project_name
  task_role_arn      = aws_iam_role.ecs_task_role.arn
  execution_role_arn = aws_iam_role.ecs_exec_role.arn
  network_mode       = "awsvpc"
  memory             = "256"

  container_definitions = jsonencode([{
    name      = "kafka",
    image     = "confluentinc/cp-kafka:latest",
    essential = true,
    portMappings = [
      {
        containerPort = 9092,
        hostPort      = 9092
      }
    ],

    environment = [
      {
        "name" : "KAFKA_ADVERTISED_LISTENERS",
        "value" : "PLAINTEXT://localhost:9092"
      },
      {
        "name" : "KAFKA_BROKER_ID",
        "value" : "1"
      },
      {
        "name" : "KAFKA_ZOOKEEPER_CONNECT",
        "value" : "zookeeper:2181"
      }
    ]

    logConfiguration = {
      logDriver = "awslogs",
      options = {
        "awslogs-region"        = var.aws_region,
        "awslogs-group"         = aws_cloudwatch_log_group.ecs.name,
        "awslogs-stream-prefix" = "kafka"
      }
    },
  }])
}

resource "aws_ecs_task_definition" "zookeeper" {
  family             = local.project_name
  task_role_arn      = aws_iam_role.ecs_task_role.arn
  execution_role_arn = aws_iam_role.ecs_exec_role.arn
  network_mode       = "awsvpc"
  memory             = "256"

  container_definitions = jsonencode([{
    name      = "zookeeper",
    image     = "confluentinc/cp-zookeeper:latest",
    essential = true,
    portMappings = [
      {
        containerPort = 2181,
        hostPort      = 2181
      }
    ],

    environment = [
      {
        "name" : "ZOOKEEPER_CLIENT_PORT",
        "value" : "2181"
      },
      {
        "name" : "ZOOKEEPER_TICK_TIME",
        "value" : "2000"
      }
    ]

    logConfiguration = {
      logDriver = "awslogs",
      options = {
        "awslogs-region"        = var.aws_region,
        "awslogs-group"         = aws_cloudwatch_log_group.ecs.name,
        "awslogs-stream-prefix" = "zookeeper"
      }
    },
  }])
}

# --- ECS Service ---

resource "aws_security_group" "ecs_task" {
  name_prefix = "ecs-task-sg-"
  description = "Allow all traffic within the VPC"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_ecs_service" "kafka" {
  name            = "kafka"
  cluster         = aws_ecs_cluster.mig.id
  task_definition = aws_ecs_task_definition.kafka.arn
  desired_count   = 1

  network_configuration {
    security_groups = [aws_security_group.ecs_task.id]
    subnets         = [for s in aws_subnet.private : s.id]
  }

  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.main.name
    base              = 1
    weight            = 100
  }

  ordered_placement_strategy {
    type  = "spread"
    field = "attribute:ecs.availability-zone"
  }

  lifecycle {
    ignore_changes = [desired_count]
  }
}

# resource "aws_ecs_task_definition" "chatservice" {
#   family             = local.project_name
#   task_role_arn      = aws_iam_role.ecs_task_role.arn
#   execution_role_arn = aws_iam_role.ecs_exec_role.arn
#   network_mode       = "awsvpc"
#   memory             = "256"

#   container_definitions = jsonencode([{
#     name      = "chatservice",
#     image     = "${aws_ecr_repository.chatservice.repository_url}:latest",
#     essential = true,
#     portMappings = [
#       {
#         containerPort = 8080,
#         hostPort      = 8001
#       }
#     ],

#     environment = [
#       { name  = "EXAMPLE",
#         value = "example"
#       }
#     ]

#     logConfiguration = {
#       logDriver = "awslogs",
#       options = {
#         "awslogs-region"        = var.aws_region,
#         "awslogs-group"         = aws_cloudwatch_log_group.ecs.name,
#         "awslogs-stream-prefix" = "chatservice"
#       }
#     },
#   }])
# }

# resource "aws_ecs_task_definition" "userservice" {
#   family             = local.project_name
#   task_role_arn      = aws_iam_role.ecs_task_role.arn
#   execution_role_arn = aws_iam_role.ecs_exec_role.arn
#   network_mode       = "awsvpc"
#   memory             = "256"

#   container_definitions = jsonencode([{
#     name      = "userservice",
#     image     = "${aws_ecr_repository.userservice.repository_url}:latest",
#     essential = true,
#     portMappings = [
#       {
#         containerPort = 8080,
#         hostPort      = 8002
#       }
#     ],

#     environment = [
#       { name  = "EXAMPLE",
#         value = "example"
#       }
#     ]

#     logConfiguration = {
#       logDriver = "awslogs",
#       options = {
#         "awslogs-region"        = var.aws_region,
#         "awslogs-group"         = aws_cloudwatch_log_group.ecs.name,
#         "awslogs-stream-prefix" = "userservice"
#       }
#     },
#   }])
# }

# # --- ECS Service ---

# resource "aws_security_group" "ecs_task" {
#   name_prefix = "ecs-task-sg-"
#   description = "Allow all traffic within the VPC"
#   vpc_id      = aws_vpc.main.id

#   ingress {
#     from_port   = 0
#     to_port     = 0
#     protocol    = "-1"
#     cidr_blocks = [aws_vpc.main.cidr_block]
#   }

#   egress {
#     from_port   = 0
#     to_port     = 0
#     protocol    = "-1"
#     cidr_blocks = ["0.0.0.0/0"]
#   }
# }

# resource "aws_ecs_service" "chatservice" {
#   name            = "chatservice"
#   cluster         = aws_ecs_cluster.mig.id
#   task_definition = aws_ecs_task_definition.chatservice.arn
#   desired_count   = 2

#   network_configuration {
#     security_groups = [aws_security_group.ecs_task.id]
#     subnets         = aws_subnet.public[*].id
#   }

#   capacity_provider_strategy {
#     capacity_provider = aws_ecs_capacity_provider.main.name
#     base              = 1
#     weight            = 100
#   }

#   ordered_placement_strategy {
#     type  = "spread"
#     field = "attribute:ecs.availability-zone"
#   }

#   lifecycle {
#     ignore_changes = [desired_count]
#   }
# }

# resource "aws_ecs_service" "userservice" {
#   name          = "chatservice"
#   cluster       = aws_ecs_cluster.mig.id
#   desired_count = 2

#   depends_on = [aws_lb_target_group.chatservice]


#   load_balancer {
#     target_group_arn = aws_lb_target_group.chatservice.arn
#     container_name   = "chatservice"
#     container_port   = 8080
#   }
# }

