provider "zeus" {
  endpoint = "http://localhost:8080"
  token    = "changeme"
}

data "zeus_port" "example" {
  id = "port-id"
}
