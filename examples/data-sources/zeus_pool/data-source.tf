provider "zeus" {
  endpoint = "http://localhost:8080"
  token    = "changeme"
}

data "zeus_pool" "example" {
  id = "pool-id"
}
