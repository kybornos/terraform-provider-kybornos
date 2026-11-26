SCHEMA = schema/kybornos.json
TOFU_DIR = examples/basic

schema:
	mkdir -p schema
	tofu -chdir=${TOFU_DIR} providers schema -json | \
		jq '.provider_schemas |= with_entries(.key = "kybornos")' \
		> ${SCHEMA}

docs-tf: schema
	tfplugindocs generate \
		--provider-name kybornos \
		--rendered-provider-name Kybornos \
		--providers-schema ${SCHEMA}

docs-hugo: schema
	tfplugindocs generate \
		--provider-name kybornos \
		--rendered-provider-name Kybornos \
		--providers-schema ${SCHEMA} \
		--website-source-dir templates-hugo \
		--rendered-website-dir docs-hugo

build:
	go build -o bin/terraform-provider-kybornos .

docs: docs-tf docs-hugo

.PHONY: schema docs docs-hugo docs-tf build
