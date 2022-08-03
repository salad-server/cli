# Build
build:
	echo "building salad-cli..."
	go build

# Build with upx (compress output)
build-prod:
	echo "building salad-cli... (with upx)"
	go build
	upx cli

# Run without building (development)
# Examples:
#   make run args="help"
#   make run args="update -s pending"
run:
	go run *.go $(args)