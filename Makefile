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
		! -name 'export_wasi_http_0_2_0_incoming_handler' \
		-exec rm -rf {} +
	componentize-go -w http-trigger -d ./wit bindings --format -o internal --pkg-name github.com/spinframework/spin-go-sdk/v3/internal
	find $$(pwd)/redis_internal/ \
		-mindepth 1 \
		-maxdepth 1 \
		-type d \
		! -name 'export_fermyon_spin_inbound_redis' \
		-exec rm -rf {} +
	componentize-go -w fermyon:spin/redis-trigger -d ./wit bindings --format -o redis_internal --pkg-name github.com/spinframework/spin-go-sdk/v3/redis_internal
