# Cleans the project directory by removing temporary files and build artifacts.
clean:
	@OUTPUTS=( "main.zip" "main" "cp.out" "tmp.yml" "tmp.json" "file.puml" "request.json" "template.json" ) sh scripts/clean-project.sh

# Tests the entire project and outputs a coverage item.
test:
	go test -coverprofile cp.out ./...

# Builds the source executable from a specified path.
build:
	GOOS=linux GOARCH=amd64 go build -o main ./handler/${DOMAIN}/main.go

# Packages the executable into a zip file with flags -9, compress better, and -r, recurse into directories.
package: build
	zip -9 -r main.zip main

# Updates an AWS λƒ; cleans afterwards.
update: package
	ENV="{\"Variables\":{}}" sh scripts/update-function.sh clean

# Creates an AWS λƒ.
create: package
	ENV="{\"Variables\":{}}" sh scripts/create-function.sh clean

# Disallow any parallelism (-j) for Make. This is necessary since some commands during the build process create
# temporary files that collide under parallel conditions.
.NOTPARALLEL:

# A phony target is one that is not really the name of a file, but rather a sequence of commands. We use this practice
# to avoid potential naming conflicts with files in the home environment but also improve performance.
.PHONY: build clean create invoke package test update

