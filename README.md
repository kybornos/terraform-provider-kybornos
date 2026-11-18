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

## Notes

- `name` and `size` changes cause VM replacement, not in-place update.
- HTTP timeouts are 30 seconds.
