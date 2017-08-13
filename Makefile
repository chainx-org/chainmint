get_vendor_deps:
	@ go get github.com/Masterminds/glide
	@ glide install

build: copy
	@ go build -i github.com/chainmint/cmd/chainmint/...

copy:
	@ mkdir -p vendor/golang.org/x/
	@ cp -r vendor/github.com/golang/crypto vendor/golang.org/x/crypto
	@ cp -r vendor/github.com/golang/net vendor/golang.org/x/net
	@ cp -r vendor/github.com/golang/text vendor/golang.org/x/text
	@ cp -r vendor/github.com/golang/tools vendor/golang.org/x/tools
	@ cp -r vendor/github.com/golang/sys vendor/golang.org/x/sys
