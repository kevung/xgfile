# Define the output directory
OUT_DIR := build

# Define the main package
MAIN_PKG := ./cmd

# Define the binary name
BINARY_NAME := xgfile

# Default target
all: build

# Build target
build:
	mkdir -p $(OUT_DIR)
	go build -o $(OUT_DIR)/$(BINARY_NAME) $(MAIN_PKG)

# Clean target
clean:
	rm -rf $(OUT_DIR)

# Phony targets
.PHONY: all build clean
