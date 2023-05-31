Param ($version)

Function Test-CommandExists {
    Param ($cmd)

    $oldPref = $ErrorActionPreference
    $ErrorActionPreference = "stop"

    Try {
        If (Get-Command $cmd) {
            return $True
        }
    }
    Catch {
        return $False
    }
    Finally {
        $ErrorActionPreference=$oldPref
    }
}

If (Test-CommandExists "go") {
    Write-Output "go found"

    go version | Tee-Object -Variable cmdOutput

    If ($cmdOutput -match 'go version go([\d\.]+) .*/.*') {
        $detectedVersion=$Matches[1]

        Write-Output "Detected Go version: $detectedVersion"

        If ($detectedVersion -eq $version) {
            Write-Output "Detected Go version matches requested version"
            return
        }

        Write-Output "Go version mismatch"
    } Else {
        Write-Output "Could not parse version from `"go version`""
    }
} Else {
    Write-Output "go not found"
}

New-Item -ItemType Directory -Force -Path C:\go-installers

$installerPath="C:\go-installers\go$version.msi"

If (-Not (Test-Path -Path $installerPath -PathType Leaf)) {
    $uri="https://storage.googleapis.com/golang/go$version.windows-amd64.msi"
    Write-Output "Downloading $uri"
    Invoke-WebRequest -Uri $uri -OutFile $installerPath
}

Write-Output "Installing go$version.msi"
msiexec /i $installerPath /q
