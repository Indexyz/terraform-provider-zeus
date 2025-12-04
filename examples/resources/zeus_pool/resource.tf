provider "zeus" {
  endpoint = "http://localhost:8080"
  token    = "changeme"
}

resource "zeus_pool" "example" {
  start   = 3232235777 # 192.168.1.1 as integer
  gateway = 3232236030 # 192.168.1.254 as integer
  size    = 16
  region  = "us-east-1"
}

data "zeus_pool" "by_id" {
  id = zeus_pool.example.id
}
