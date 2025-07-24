$reportDir = "..\security-reports"

if (!(Test-Path $reportDir)) {
    New-Item -ItemType Directory -Path $reportDir
}

Get-ChildItem -Directory | ForEach-Object {
    $name = $_.Name
    Write-Host "?? Skeniram $name..."
    $outputFile = Join-Path $reportDir ("report-" + $name + ".txt")
    $targetPath = Join-Path "." $name
    trivy fs $targetPath --scanners vuln,misconfig --severity HIGH,CRITICAL --output $outputFile
}