foreach ($line in Get-Content 'C:\Users\Jay Chinthrajah\Workspaces\ai-upskill\.env') {
    if ($line -match '^([^#][^=]*)=(.+)$') {
        $k = $Matches[1].Trim()
        $v = $Matches[2].Trim()
        if (($v.StartsWith("'") -and $v.EndsWith("'")) -or ($v.StartsWith('"') -and $v.EndsWith('"'))) {
            $v = $v.Substring(1, $v.Length - 2)
        }
        [System.Environment]::SetEnvironmentVariable($k, $v, 'Process')
    }
}
$val = [System.Environment]::GetEnvironmentVariable('NOTEBOOKLM_AUTH_JSON', 'Process')
Write-Host ("Length: " + $val.Length)
Write-Host ("First char: '" + $val[0] + "'")
try {
    $parsed = $val | ConvertFrom-Json
    Write-Host ("JSON valid, cookies: " + $parsed.cookies.Count)
} catch {
    Write-Host ("JSON parse error: " + $_)
}
