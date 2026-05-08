# Load .env and verify NOTEBOOKLM_AUTH_JSON
$envFile = "C:\Users\Jay Chinthrajah\Workspaces\ai-upskill\.env"
$lines = Get-Content $envFile
foreach ($line in $lines) {
    if ($line -match '^([^=]+)=(.+)$') {
        $k = $Matches[1].Trim()
        $v = $Matches[2].Trim()
        [System.Environment]::SetEnvironmentVariable($k, $v, 'Process')
    }
}
$val = [System.Environment]::GetEnvironmentVariable('NOTEBOOKLM_AUTH_JSON', 'Process')
if ($val) {
    Write-Output "SET (length: $($val.Length))"
} else {
    Write-Error "NOTEBOOKLM_AUTH_JSON not set after loading .env"
    exit 1
}
