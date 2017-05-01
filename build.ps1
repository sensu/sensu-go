param (
  [string] $cmd = $null
)

function install_deps
{
  go get github.com/axw/gocov/gocov
  go get gopkg.in/alecthomas/gometalinter.v1
  go get github.com/gordonklaus/ineffassign
  go get github.com/jgautheron/goconst/cmd/goconst
  go get -u github.com/golang/lint/golint
}

function build_binary
{
  # build binary
}

function build_commands
{
  echo "Running build..."
}

function test_commands
{
  echo "Running tests..."

  gometalinter.v1 --vendor --disable-all --enable=vet --enable=vetshadow --enable=golint --enable=ineffassign --enable=goconst --tests ./...
}

If ($cmd -eq "deps") {
  install_deps
}
ElseIf ($cmd -eq "unit") {
  test_commands

  If ($LASTEXITCODE -ne 0) {
    echo "Linting failed..."
    exit 1
  }

  echo "" > "coverage.txt"
  $packages = go list ./... | Select-String -pattern "testing", "vendor" -notMatch
  ForEach($pkg in $packages) {
    go test -timeout=60s -v -coverprofile="profile.out" -covermode=atomic $pkg
    If (Test-Path "profile.out") {
      cat "profile.out" >> "coverage.txt"
      rm "profile.out"
    }
  }
}
ElseIf ($cmd -eq "build") {
  build_commands
}
Else {
  install_deps
  test_commands
  build_commands
  e2e_commands
}
