function cleanup() {
    rm -f export
}
trap cleanup EXIT

export COOKIE=""
export CSRF_TOKEN=""

# Compile Go
GO111MODULE=on GOGC=off go build -mod=vendor -v -o export export.go
./export

