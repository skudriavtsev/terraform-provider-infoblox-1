terraform {
  # Required providers block for Terraform v0.14
  required_providers {
    infoblox = {
      source  = "terraform-providers/infoblox"
      version = "~> 1.1.0"
    }
  }
}
