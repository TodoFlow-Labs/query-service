
# Makefile for projection-worker
# This Makefile is used to build and run the projection-worker service.

.PHONY: run
run:
	# Run the projection-worker service with the specified NATS URL and Bleve index path.
	# The log level is set to debug for detailed logging.
	# The Bleve index path is set to /data/index.bleve.
	# The NATS URL is set to nats://nats:4222.
	# The command is run in the current directory.
	go run cmd/main.go \
	--http-addr ":8082" \
	--bleve-index-path "../projection-worker/data/index.bleve" \
	--log-level debug 