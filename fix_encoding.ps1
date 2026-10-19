$content = Get-Content "J:\ScarScript\docs\site\index.html" -Raw -Encoding UTF8

# Fix all the garbled UTF-8 sequences using literal strings
$content = $content -replace "Pakistan…'s", "Pakistan's"
$content = $content -replace "High…Performance", "High-Performance"
$content = $content -replace "runtime…with", "runtime — with"
$content = $content -replace "nothing…you", "nothing — you"
$content = $content -replace "VM…just raw", "VM — just raw"
$content = $content -replace "interpolation…feels", "interpolation — feels"
$content = $content -replace "binaries…deliver", "binaries — deliver"
$content = $content -replace "1\.00… \(baseline\)", "1.00× (baseline)"
$content = $content -replace "1\.06… faster", "1.06× faster"
$content = $content -replace "5\.8… slower", "5.8× slower"
$content = $content -replace "23… slower", "23× slower"
$content = $content -replace "6…23… faster", "6–23× faster"
$content = $content -replace "compilation…just", "compilation — just"
$content = $content -replace "platform…no Go", "platform — no Go"
$content = $content -replace "inference…just use", "inference — just use"
$content = $content -replace "time…full reference", "time — full reference"
$content = $content -replace "Scarfaze… <strong>Scarfaze</strong>", "Scarfaze — <strong>Scarfaze</strong>"

# Meta tags
$content = $content -replace "ScarScript… Pakistan", "ScarScript — Pakistan"
$content = $content -replace "ScarScript… High-Performance", "ScarScript — High-Performance"

Set-Content "J:\ScarScript\docs\site\index.html" -Value $content -Encoding UTF8
Write-Host "Done"