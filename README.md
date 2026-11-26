# Kybornos OpenTofu/Terraform provider

This is a minimal Terraform provider for the Kybornos Gateway HTTP API.

## API endpoints

- `POST /v1/vms/deploy` with `{"name":"web01","size":"small"}` → JSON `{"id":"...","name":"web01","size":"small"}`
- `GET /v1/vms/:id` → JSON `{"id":"...","name":"web01","size":"small"}` or HTTP 404
- `POST /v1/vms/:id/destroy` → HTTP 2xx or 404
- `Authorization: Bearer <KYBORNOS_API_KEY>`

## Local build

```sh
go mod tidy
go test ./...
mkdir -p bin
go build -o bin/terraform-provider-kybornos .
```

For local development, copy `examples/basic/dev.tfrc.example` to a file and substitute the absolute path to the `bin` directory. Then:

```sh
export TF_CLI_CONFIG_FILE=/absolute/path/to/dev.tfrc
export KYBORNOS_API_KEY='your-key'
cd examples/basic
tofu plan
tofu apply
```

Skip `tofu init` as the provider is not published yet.

## Import an existing VM

```sh
tofu import kybornos_vm.web '<vm-id>'
```

## Example TF state

```sh
$ tofu show
# kybornos_vm.web[0]:
resource "kybornos_vm" "web" {
    cpu        = 1
    disk       = 20480
    hostname   = "web0"
    id         = "01a2172d-456d-7d65-b54e-50f3188f537f"
    interfaces = [
        {
            ip4     = "10.0.11.13"
            lladdr  = "4e:2a:4b:05:94:ac"
            primary = true
            public  = false
        },
    ]
    memory     = 1024
    public     = false
    size       = "small"
    vm_name    = "A0424BE62857.web0"
}


Outputs:

ip_addresses = [
    [
        "10.0.11.13",
    ],
]
vm_ids = [
    "01a2172d-456d-7d65-b54e-50f3188f537f",
]
```

## Notes

- `hostname` and `size` changes cause VM replacement, not in-place update.
- HTTP timeouts are 30 seconds.
