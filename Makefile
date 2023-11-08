
build:
	go build -o vpa .

install:
	go install .

release:
	gh release create $(TAG) -t $(TAG)

check:
	@echo "Checking...\n"
	gocyclo -over 15 . || echo -n ""
	@echo ""
	golint -min_confidence 0.21 -set_exit_status ./...
	@echo "\nAll ok!"