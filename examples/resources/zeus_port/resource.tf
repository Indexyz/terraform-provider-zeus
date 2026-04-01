provider "zeus" {
  endpoint = "http://localhost:8080"
  token    = "changeme"
}

resource "zeus_assign" "vm" {
  region = ["us-east-1"]
  host   = "node-1"
  key    = "vm-123"
  type   = "vm"
}

resource "zeus_port" "ssh" {
  assign_id   = zeus_assign.vm.id
  scope_host  = "node-1"
  target_port = 22
  service     = "ssh"
}

data "zeus_port" "by_id" {
  id = zeus_port.ssh.id
}
