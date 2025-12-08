plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

plugin "aws" {
  enabled = true
  version = "0.35.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

config {
  # Ignore downloaded modules
  ignore_module = {
    "terraform-aws-modules/vpc/aws"  = true
    "terraform-aws-modules/eks/aws"  = true
    "terraform-aws-modules/iam/aws"  = true
  }
}

# Naming conventions
rule "terraform_naming_convention" {
  enabled = true
}

# Require descriptions for variables
rule "terraform_documented_variables" {
  enabled = true
}

# Require descriptions for outputs
rule "terraform_documented_outputs" {
  enabled = true
}

# Standard module structure
rule "terraform_standard_module_structure" {
  enabled = false  # Disable for single-directory setup
}
