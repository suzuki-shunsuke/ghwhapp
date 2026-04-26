variable "api_gateway_name" {
  type    = string
  default = "ghwhapp"
}

variable "secretsmanager_secret_name_main" {
  type    = string
  default = "ghwhapp"
}

variable "lambda_architecture" {
  type        = string
  description = "Lambda Architecture"
  default     = "arm64"
}

variable "zip_path" {
  type        = string
  description = "Lambda Zip File Path"
  default     = "ghwhapp_linux_arm64.zip"
}

variable "function_name" {
  type        = string
  description = "Lambda Function Name"
  default     = "ghwhapp"
}

variable "lambda_role_name" {
  type        = string
  description = "Lambda Role Name"
  default     = "ghwhapp"
}

variable "lambda_role_path" {
  type        = string
  description = "Lambda Role Path"
  default     = "/service-role/"
}
