param (
    [string] $cmd = $null
)

$REPO_PATH = "github.com/sensu/sensu-go"

$VERSION_CMD = "go run ./version/cmd/version/version.go"

$env_commands = go env
ForEach ($env_cmd in $env_commands) {
    $env_str = $env_cmd -replace "set " -replace ""
    $env = $env_str.Split("=")
    $ps_cmd = "`$env:$($env[0]) = ""$($env[1])"""
    Invoke-Expression $ps_cmd
}

$RACE = ""

function set_race_flag
{
    If ($env:GOARCH -eq "amd64") {
        $RACE = "-race"
    }
}

switch ($env:GOOS)
{
    "darwin" {set_race_flag}
    "freebsd" {set_race_flag}
    "linux" {set_race_flag}
    "windows" {set_race_flag}
}

function build_tool_binary([string]$goos, [string]$goarch, [string]$bin, [string]$subdir)
{
    $outfile = "target/$goos-$goarch/$subdir/$bin.exe"
    $env:GOOS = $goos
    $env:GOARCH = $goarch
    go build -o $outfile "$REPO_PATH/$subdir/$bin/..."
    If ($LASTEXITCODE -ne 0) {
        echo "Failed to build $outfile..."
        exit 1
    }

    return $outfile
}

function cmd_name_map([string]$cmd)
{
    switch ($cmd)
    {
        "backend" {
            return "sensu-backend"
        }
        "agent" {
            return "sensu-agent"
        }
        "cli" {
            return "sensuctl"
        }
    }
}

function build_binary([string]$goos, [string]$goarch, [string]$bin, [string]$cmd_name)
{
    $outfile = "target/$goos-$goarch/$cmd_name.exe"
    $env:GOOS = $goos
    $env:GOARCH = $goarch

    $build_type = &"go" "run" "./version/cmd/version/version.go" "-t" | Out-String
    $build_type = $build_type -replace "`n","" -replace "`r",""
    $version = &"go" "run" "./version/cmd/version/version.go" "-v" | Out-String
    $build_date = Get-Date -format "yyyy'-'MM'-'dd'T'T'-'Z"
    $build_sha = (git rev-parse HEAD) | Out-String

    $version_pkg = "github.com/sensu/sensu-go/version"
    $ldflags = "-X $version_pkg.Version=$version"
    $ldflags = $ldflags + " -X $version_pkg.BuildDate=$build_date"
    $ldflags = $ldflags + " -X $version_pkg.BuildSHA=$build_sha"

    If ($build_type -ne "nightly" -And $build_type -ne "stable") {
        $prerelease = &"go" "run" "./version/cmd/version/version.go" "-p" | Out-String
        $ldflags = $ldflags + " -X $version_pkg.PreReleaseIdentifier=$prerelease"
    }

    $main_pkg = "cmd/$cmd_name"

    go build -ldflags "$ldflags" -o $outfile "$REPO_PATH/$main_pkg"
    If ($LASTEXITCODE -ne 0) {
        echo "Failed to build $outfile..."
        exit 1
    }

    return $outfile
}

function build_tool([string]$bin, [string]$subdir)
{
    If (!(Test-Path -Path "bin/$subdir")) {
        New-Item -ItemType directory -Path "bin/$subdir" | out-null
    }

    echo "Building $subdir/$bin for $env:GOOS-$env:GOARCH"
    $out = build_tool_binary $env:GOOS $env:GOARCH $bin $subdir
    Remove-Item -Path "bin/$(Split-Path -Leaf $out)" -EA SilentlyContinue
    cp $out bin/$subdir
    dir bin
    dir bin/$subdir
}

function build_commands
{
    echo "Running build..."

    ForEach ($bin in "agent","backend","cli") {
        build_command $bin
    }
}

function build_agent
{
    build_command "agent"
}

function build_backend
{
    build_dashboard
    build_command "backend"
}

function build_cli
{
    build_command "cli"
}

function build_dashboard
{
    go generate ./dashboard
}

function build_command([string]$bin)
{
    $cmd_name = cmd_name_map $bin

    If (!(Test-Path -Path "bin")) {
        New-Item -ItemType directory -Path "bin" | out-null
    }

    echo "Building $bin for $env:GOOS-$env:GOARCH"
    $out = build_binary $env:GOOS $env:GOARCH $bin $cmd_name
    Remove-Item -Path "bin/$(Split-Path -Leaf $out)" -EA SilentlyContinue
    cp $out bin
}

function linter_commands
{
    echo "Running linter..."

    go get gopkg.in/alecthomas/gometalinter.v2
    go get github.com/gordonklaus/ineffassign
    go get github.com/jgautheron/goconst/cmd/goconst
    go get honnef.co/go/tools/cmd/megacheck

    megacheck $(go list ./... | grep -v dashboardd | grep -v agent/assetmanager | grep -v scripts)
    If ($LASTEXITCODE -ne 0) {
        echo "Linting failed..."
        exit 1
    }

    gometalinter.v2 --vendor --disable-all --enable=vet --linter='vet:go tool vet -composites=false {paths}:PATH:LINE:MESSAGE' --enable=ineffassign --enable=goconst --tests ./...
    If ($LASTEXITCODE -ne 0) {
        echo "Linting failed..."
        exit 1
    }
}

function unit_test_commands
{
    echo "Running unit tests..."

    go test -timeout=60s $(go list ./... | Select-String -pattern "scripts", "testing", "vendor" -notMatch)
    If ($LASTEXITCODE -ne 0) {
        echo "Unit testing failed..."
        exit 1
    }
}

function integration_test_commands
{
    echo "Running integration tests..."

    go test -timeout=200s -tags=integration $(go list ./... | Select-String -pattern "scripts", "testing", "vendor" -notMatch)
    If ($LASTEXITCODE -ne 0) {
        echo "Integration testing failed..."
        exit 1
    }
}

function e2e_commands
{
    echo "Running e2e tests..."

    go test $REPO_PATH/testing/e2e
    If ($LASTEXITCODE -ne 0) {
        echo "e2e testing failed..."
        exit 1
    }
}

function wait_for_appveyor_jobs {
    if ($env:APPVEYOR_JOB_NUMBER -ne 1) { return }

    write-host "Waiting for other jobs to complete"

    $headers = @{
        "Authorization" = "Bearer $env:APPVEYOR_API_TOKEN"
        "Content-type" = "application/json"
    }

    [datetime]$stop = ([datetime]::Now).AddMinutes($env:TIME_OUT_MINS)
    [bool]$success = $false

    while(!$success -and ([datetime]::Now) -lt $stop) {
        $project = Invoke-RestMethod -Uri "https://ci.appveyor.com/api/projects/$env:APPVEYOR_ACCOUNT_NAME/$env:APPVEYOR_PROJECT_SLUG" -Headers $headers -Method GET
        $success = $true
        $project.build.jobs | foreach-object {if (($_.jobId -ne $env:APPVEYOR_JOB_ID) -and ($_.status -ne "success")) {$success = $false}; $_.jobId; $_.status}
        if (!$success) {Start-sleep 5}
    }

    if (!$success) {throw "Test jobs were not finished in $env:TIME_OUT_MINS minutes"}
}

function build_package([string]$package, [string]$arch)
{
    echo "Building $package MSI"

    rm *.wixobj
    rm *.wixpdb

    $package_base_name = "sensu-$package"
    If ($package -eq "cli") {
        $package_base_name = "sensuctl"
    }

    $version = &"go" "run" "./version/cmd/version/version.go" "-v" | Out-String
    $iteration = &"go" "run" "./version/cmd/version/version.go" "-i" | Out-String
    $full_version = "${version}.${iteration}"
    $msi_filename = "${package_base_name}_${full_version}_${arch}.msi"

    echo "Package version: $version"
    echo "Package iteration: $iteration"
    echo "Package full version: $full_version"

    echo "Wix path: $Env:WIX"
    $Env:Path += ";$($Env:WIX)\\bin"

    candle.exe -arch $arch packaging/wix/$package/*.wxs -ext WixUtilExtension -dProjectDir="${PWD}" -dVersionNumber="${full_version}"
    light.exe -nologo -dcl:high -ext WixUIExtension -ext WixUtilExtension -ext WixNetFxExtension *.wixobj -cultures:en-us -loc packaging/wix/$package/Product_en-us.wxl -o $msi_filename

    ls

    Push-AppveyorArtifact $msi_filename
}

If ($cmd -eq "build") {
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
ElseIf ($cmd -eq "docker") {
    # no-op for now
}
ElseIf ($cmd -eq "e2e") {
    build_commands
    e2e_commands
}
ElseIf ($cmd -eq "lint") {
    linter_commands
}
ElseIf ($cmd -eq "quality") {
    unit_test_commands
}
ElseIf ($cmd -eq "unit") {
    unit_test_commands
}
ElseIf ($cmd -eq "integration") {
    integration_test_commands
}
ElseIf ($cmd -eq "wait_for_appveyor_jobs") {
    If (($env:APPVEYOR_REPO_TAG -eq $true) -and ($env:MSI_BUILDER -eq $true)) {
        $env:GOARCH = "amd64"
        build_command "agent"

        $env:GOARCH = "386"
        build_command "agent"

        wait_for_appveyor_jobs

        build_package "agent" "x64"
        build_package "agent" "x86"
    }
}
Else {
    unit_test_commands
    integration_test_commands
    build_commands
    e2e_commands
}
