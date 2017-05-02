param (
  [string] $cmd = $null
)

$REPO_PATH = "github.com/sensu/sensu-go"

# source in the environment variables from `go env`
$env_commands = go env
ForEach ($env_cmd in $env_commands) {
  $env_str = $env_cmd -replace "set " -replace ""
  $env = $env_str.Split("=")
  $ps_cmd = "`$env:$($env[0]) = ""$($env[1])"""
  Invoke-Expression $ps_cmd
}

function install_deps
{
  go get github.com/axw/gocov/gocov
  go get gopkg.in/alecthomas/gometalinter.v1
  go get github.com/gordonklaus/ineffassign
  go get github.com/jgautheron/goconst/cmd/goconst
  go get -u github.com/golang/lint/golint
}

function build_binary([string]$goos, [string]$goarch, [string]$bin)
{
  $outfile = "target/$goos-$goarch/sensu-$bin.exe"
  $env:GOOS = $goos
  $env:GOARCH = $goarch
  go build -o $outfile "$REPO_PATH/$bin/cmd/..."

  return $outfile
}

function build_commands
{
  echo "Running build..."

  ForEach ($bin in "agent","backend","cli") {
    build_command $bin
  }
}

function build_command([string]$bin)
{
  If (!(Test-Path -Path "bin")) {
    New-Item -ItemType directory -Path "bin" | out-null
  }

  echo "Building $bin for $env:GOOS-$env:GOARCH"
  $out = build_binary $env:GOOS $env:GOARCH $bin
  Remove-Item -Path "bin/$(Split-Path -Leaf $out)" -EA SilentlyContinue
  cp $out bin
}

function test_commands
{
  echo "Running tests..."

  gometalinter.v1 --vendor --disable-all --enable=vet --enable=vetshadow --enable=golint --enable=ineffassign --enable=goconst --tests ./...
  If ($LASTEXITCODE -ne 0) {
    echo "Linting failed..."
    exit 1
  }

  echo "" > "coverage.txt"
  $packages = go list ./... | Select-String -pattern "testing", "vendor" -notMatch
  ForEach ($pkg in $packages) {
    go test -timeout=60s -v -coverprofile="profile.out" -covermode=atomic $pkg
    If (Test-Path "profile.out") {
      cat "profile.out" >> "coverage.txt"
      rm "profile.out"
    }
  }
}

function e2e_commands
{
  echo "Running e2e tests..."

  go test -v $REPO_PATH/testing/e2e
}

If ($cmd -eq "deps") {
  install_deps
}
ElseIf ($cmd -eq "unit") {
  test_commands
}
ElseIf ($cmd -eq "build") {
  build_commands
}
ElseIf ($cmd -eq "build_agent") {
  build_command "agent"
}
ElseIf ($cmd -eq "build_backend") {
  build_command "backend"
}
ElseIf ($cmd -eq "build_cli") {
  build_command "cli"
}
Else {
  install_deps
  test_commands
  build_commands
  e2e_commands
}
