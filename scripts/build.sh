#!/bin/bash

export CGO_ENABLED=0

VERSION=${1:-develop}
PROJECT=koyeb-cli
OUT_DIR=./build_output
LINUX="$OUT_DIR/$PROJECT-$VERSION-linux-x86_64"
LINUX_386="$OUT_DIR/$PROJECT-$VERSION-linux-386"
DARWIN="$OUT_DIR/$PROJECT-$VERSION-darwin-x86_64"
WINDOWS="$OUT_DIR/$PROJECT-$VERSION-windows-x86_64.exe"
WINDOWS_386="$OUT_DIR/$PROJECT-$VERSION-windows-386.exe"
LDFLAGS="-X github.com/koyeb/koyeb-cli/pkg/koyeb.Version=$VERSION"

echo Building $LINUX
GOOS=linux GOARCH=amd64 go build -a -ldflags="$LDFLAGS" -installsuffix nocgo -o "$LINUX" cmd/koyeb/koyeb.go
echo Building $LINUX_386
GOOS=linux GOARCH=386 go build -a -ldflags="$LDFLAGS" -installsuffix nocgo -o "$LINUX_386" cmd/koyeb/koyeb.go
echo Building $DARWIN
GOOS=darwin GOARCH=amd64 go build -a -ldflags="$LDFLAGS" -installsuffix nocgo -o "$DARWIN" cmd/koyeb/koyeb.go
echo Building $WINDOWS
GOOS=windows GOARCH=amd64 go build -a -ldflags="$LDFLAGS" -installsuffix nocgo -o "$WINDOWS" cmd/koyeb/koyeb.go
echo Building $WINDOWS_386
GOOS=windows GOARCH=386 go build -a -ldflags="$LDFLAGS" -installsuffix nocgo -o "$WINDOWS_386" cmd/koyeb/koyeb.go




