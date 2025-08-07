# fal Terraform Provider

The [fal Terraform Provider](https://registry.terraform.io/providers/fal-ai/fal/latest/docs) offers convenient access to [fal's APIs](https://docs.fal.ai/serverless) from Terraform.

## Requirements

You will need to [install uv](https://docs.astral.sh/uv/getting-started/installation/) into your Terraform host environment.

If you're running Terraform in GitHub Actions, an example of this might look like:
```yaml
name: Example

jobs:
   terraform-uv-example:
      name: python
      runs-on: ubuntu-latest

      steps:
         - uses: actions/checkout@v4

         - name: Install uv
           uses: astral-sh/setup-uv@v6

         - name: Install terraform
           uses: hashicorp/setup-terraform@v3
      
         - run: terraform init
```

## Usage

Add the following to your `main.tf` file:
```terraform
terraform {
  required_providers {
    fal = {
      source = "fal-ai/fal"
      version = "~> 1"
    }
  }
}

provider "fal" {
  fal_key = "<fal key here>"
}

resource "fal_app" "sana_app" {
  entrypoint = "fal_demos/image/sana.py"
  git = {
    url = "https://github.com/fal-ai-community/fal-demos.git"
  }
}
```

Initialize your project by running `terraform init` in the directory.

You can refer to the full documentation on [the Terraform Registry](https://registry.terraform.io/providers/fal-ai/fal/latest/docs).
