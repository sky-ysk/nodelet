BINARY_NAME=syncserver
BUILD_DIR=bin
MAIN_FILE=main.go
LDFLAGS=-s -w
VERSION=1.0.0

.PHONY: check-dir
check-dir:
	@mkdir -p $(BUILD_DIR)

# 清理
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

# 构建
.PHONY: build
build: check-dir
	@echo "Building application..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags "$(LDFLAGS)" $(MAIN_FILE)
	@echo "Build complete. Output: "$(BUILD_DIR)/$(BINARY_NAME)""
