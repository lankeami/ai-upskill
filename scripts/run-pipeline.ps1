# Daily pipeline runner for Windows (loads .env, then runs each step)
param([string]$Date = (Get-Date -Format "yyyy-MM-dd"))

$Root = "C:\Users\Jay Chinthrajah\Workspaces\ai-upskill"
Set-Location $Root

# Load .env
$envFile = Join-Path $Root ".env"
if (Test-Path $envFile) {
    foreach ($line in Get-Content $envFile) {
        if ($line -match '^([^#][^=]*)=(.+)$') {
            $k = $Matches[1].Trim()
            $v = $Matches[2].Trim()
            # Strip surrounding single or double quotes
            if (($v.StartsWith("'") -and $v.EndsWith("'")) -or ($v.StartsWith('"') -and $v.EndsWith('"'))) {
                $v = $v.Substring(1, $v.Length - 2)
            }
            [System.Environment]::SetEnvironmentVariable($k, $v, 'Process')
        }
    }
    Write-Host "Loaded .env"
}

Write-Host "DATE: $Date"

# Step 0: git pull
Write-Host ""
Write-Host "=== Step 0: Sync main ==="
git pull origin main
if ($LASTEXITCODE -ne 0) { Write-Error "git pull failed"; exit 1 }

# Step 1: Build Go CLI
Write-Host ""
Write-Host "=== Step 1: Build Go CLI ==="
go build -o ai-report.exe ./cmd/ai-report
if ($LASTEXITCODE -ne 0) { Write-Error "go build failed"; exit 1 }
Write-Host "Build OK"

# Step 2: Generate report
Write-Host ""
Write-Host "=== Step 2: Generate Report ==="
$reportPath = Join-Path $Root "reports\$Date.md"
if (Test-Path $reportPath) {
    Write-Host "Report already exists for $Date -- skipping"
    $reportGenerated = $false
} else {
    .\ai-report.exe generate --date $Date
    if ($LASTEXITCODE -ne 0) { Write-Error "Report generation failed"; exit 1 }
    if (-not (Test-Path $reportPath)) { Write-Error "Report file not created"; exit 1 }
    Write-Host "Report generated: $reportPath"
    $reportGenerated = $true
}

# Step 3: Generate audio
Write-Host ""
Write-Host "=== Step 3: Generate Audio ==="
$podcastPath = Join-Path $Root "podcasts\$Date.mp3"
if (Test-Path $podcastPath) {
    Write-Host "Audio already exists for $Date -- skipping"
    $audioGenerated = $false
} else {
    # Step 3a: Start
    python scripts\generate-podcast.py start --date $Date --media-type audio
    if ($LASTEXITCODE -ne 0) { Write-Error "Audio start failed"; exit 1 }
    Write-Host "Audio generation started. Polling for completion..."

    # Step 3b: Poll
    $maxPolls = 40
    $poll = 0
    $done = $false
    while ($poll -lt $maxPolls) {
        $poll++
        Start-Sleep -Seconds 30
        $pollOut = python scripts\generate-podcast.py poll 2>&1
        $pollExit = $LASTEXITCODE
        Write-Host "Poll ${poll}/${maxPolls}: $pollOut"
        if ($pollExit -eq 0) { $done = $true; break }
        if ($pollExit -eq 2) { Write-Error "Audio generation failed: $pollOut"; exit 1 }
    }
    if (-not $done) { Write-Error "Audio generation timed out after 20 minutes"; exit 1 }

    # Step 3c: Download
    python scripts\generate-podcast.py download
    if ($LASTEXITCODE -ne 0) { Write-Error "Audio download failed"; exit 1 }
    if (-not (Test-Path $podcastPath)) { Write-Error "Podcast file not created"; exit 1 }
    Write-Host "Audio downloaded: $podcastPath"
    $audioGenerated = $true
}

# Step 4: Publish GitHub Release
Write-Host ""
Write-Host "=== Step 4: GitHub Release ==="
gh release view "podcast-$Date" 2>&1 | Out-Null
if ($LASTEXITCODE -eq 0) {
    Write-Host "Release already exists for $Date -- skipping"
    $releaseCreated = $false
} else {
    gh release create "podcast-$Date" "podcasts\$Date.mp3" --title "Podcast -- $Date" --notes "Audio podcast for AI Daily Report $Date"
    if ($LASTEXITCODE -ne 0) { Write-Error "Release creation failed"; exit 1 }
    Write-Host "Release created: podcast-$Date"
    $releaseCreated = $true
}

$podcastUrl = "https://github.com/lankeami/ai-upskill/releases/download/podcast-$Date/$Date.mp3"

# Inject podcast_url into front matter if missing
$fmCheck = Select-String -Path $reportPath -Pattern "podcast_url:" -Quiet
if ($fmCheck) {
    Write-Host "Front matter already has podcast_url -- skipping injection"
    $fmUpdated = $false
} else {
    $content = Get-Content $reportPath -Raw
    $parts = $content -split '---', 3
    if ($parts.Length -ge 3) {
        $parts[1] = $parts[1].TrimEnd() + "`npodcast_url: `"$podcastUrl`"`n"
        $parts -join '---' | Set-Content $reportPath -NoNewline
        Write-Host "Injected podcast_url into $reportPath"
        $fmUpdated = $true
    } else {
        Write-Error "Could not parse front matter in $reportPath"
        exit 1
    }
}

# Step 5: Commit and push
Write-Host ""
Write-Host "=== Step 5: Commit and Push ==="
$gitStatus = git status --porcelain
if (-not $gitStatus) {
    Write-Host "Nothing to commit"
    $committed = $false
} else {
    git add "reports\$Date.md"
    if ($reportGenerated) {
        git commit -m "chore: daily AI report for $Date"
    } else {
        git commit -m "chore: add podcast URL to $Date report"
    }
    git push
    if ($LASTEXITCODE -ne 0) { Write-Error "git push failed"; exit 1 }
    Write-Host "Pushed to main"
    $committed = $true
}

# Summary
Write-Host ""
Write-Host "=== Done ==="
if ($reportGenerated) { Write-Host "Report:   generated" } else { Write-Host "Report:   already existed" }
if ($audioGenerated) { Write-Host "Audio:    generated" } else { Write-Host "Audio:    already existed" }
if ($releaseCreated) { Write-Host "Release:  published" } else { Write-Host "Release:  already existed" }
if ($committed) { Write-Host "Commit:   pushed" } else { Write-Host "Commit:   nothing to commit" }
