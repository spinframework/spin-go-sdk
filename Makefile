.PHONY: test-integration
test-integration:
	go test -v -count=1 .

.PHONY: build-examples
build-examples:
	for x in examples/*; do echo $$x && (cd $$x && spin build) || exit 1; done

.PHONY: regenerate-bindings
regenerate-bindings:
	find $$(pwd)/internal/ \
		-mindepth 1 \
		-maxdepth 1 \
		-type d \
		! -name 'db' \
		! -name 'export_fermyon_spin_inbound_redis' \
		! -name 'export_wasi_http_0_2_0_incoming_handler' \
		-exec rm -rf {} +
	componentize-go \
		--ignore-toml-files \
		-w "fermyon:spin/http-trigger@3.0.0" \
		-w "fermyon:spin/redis-trigger" \
		-d ./wit \
		bindings \
		--format \
		-o internal \
		--pkg-name github.com/spinframework/spin-go-sdk/v3/internal
