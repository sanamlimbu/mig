output "mig_server_public_dns" {
  value       = aws_instance.mig.public_dns
  description = "The public DNS of the Mig Server EC2 instance."
}
