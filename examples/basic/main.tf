terraform {
  required_providers {
    kybornos = {
      source = "registry.opentofu.org/kybornos/kybornos"
    }
  }
}

// Accessible only over Wireguard
provider "kybornos" {
  endpoint = "https://api0.secure.t1ops.net:8443"
}

resource "kybornos_vm" "web" {
  count    = 2
  hostname = "web${count.index}"
  size     = "small"
  public   = false
  autoinstall = {
    timezone = "Europe/Athens"
    username = "dave"
  }
  srv_tags = [
    {
      name           = "http"
      port           = 80
      weight         = 10
      priority       = 10
      protocol       = "tcp"
      ttl            = 60
      tls            = false
      alt_names      = ["www"]
      ingress_domain = "example.tld"
    }
  ]
}

output "vm_ids" {
  value = kybornos_vm.web[*].id
}

output "ip_addresses" {
  value = kybornos_vm.web[*].interfaces[*].ip4
}
