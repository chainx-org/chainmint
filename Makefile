get_vendor_deps:
	@ go get github.com/Masterminds/glide
	@ glide install

build:
	@ go build -i github.com/chainmint/cmd/chainmint/...
