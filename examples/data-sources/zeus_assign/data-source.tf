provider "zeus" {
  endpoint = "http://localhost:8080"
  token    = "changeme"
}

data "zeus_assign" "example" {
  id = "assign-id"
}
