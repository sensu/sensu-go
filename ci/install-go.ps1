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

Write-Output "Installing Go $version"

$uri="https://storage.googleapis.com/golang/go$version.windows-amd64.msi"

Write-Output "Downloading $uri"

Invoke-WebRequest -Uri $uri -OutFile "C:\go.msi"

Write-Output "Installing go.msi"

msiexec /i C:\go.msi /q
