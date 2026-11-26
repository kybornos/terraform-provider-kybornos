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
  count = 1
  hostname = "web${count.index}"
  size = "small"
  public = false
}

output "vm_ids" {
  value = kybornos_vm.web[*].id
}

output "ip_addresses" {
    value = kybornos_vm.web[*].interfaces[*].ip4
}
