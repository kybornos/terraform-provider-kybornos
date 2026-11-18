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
  name = "web01"
  size = "small"
}

output "vm_id" {
  value = kybornos_vm.web.id
}
