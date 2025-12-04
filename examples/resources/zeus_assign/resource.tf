provider "zeus" {
  endpoint = "http://localhost:8080"
  token    = "changeme"
}

resource "zeus_assign" "example" {
  region = ["us-east-1"]
  host   = "host-1"
  key    = "vm-123"
  type   = "vm"
  data   = { env = "dev" }
}

data "zeus_assign" "by_id" {
  id = zeus_assign.example.id
}
