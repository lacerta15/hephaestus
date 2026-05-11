# Hephaestus to GitHub - one-shot push script
#
# Usage:
#   .\push-to-github.ps1 -RepoUrl https://github.com/lacerta15/hephaestus.git

param(
  [Parameter(Mandatory=$true)]
  [string]$RepoUrl,

  [string]$UserName  = "DSI",
  [string]$UserEmail = "cloudera.dsi@gmail.com",
  [string]$Branch    = "main"
)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "======================================================="
Write-Host " Hephaestus to GitHub push"
Write-Host "======================================================="
Write-Host ""

# 0. Check git
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
  Write-Host "ERROR: Git is not installed. Install from https://git-scm.com/download/win"
  exit 1
}
Write-Host "[OK] Git found: $(git --version)"

# 1. Remove corrupt .git if exists
if (Test-Path .git) {
  Write-Host "[..] Removing existing .git directory"
  Remove-Item -Recurse -Force .git
}

# 2. Init repo
Write-Host "[..] git init"
git init -b $Branch | Out-Null
git config user.name  $UserName
git config user.email $UserEmail
git config core.autocrlf input
Write-Host "[OK] Local repo initialised on branch: $Branch"

# 3. Stage everything
Write-Host "[..] git add ."
git add .
$staged = (git diff --cached --name-only | Measure-Object -Line).Lines
Write-Host "[OK] $staged files staged"

# 4. Safety scan for secrets
$leaks = git diff --cached --name-only | Select-String -Pattern "\.pem$|\.key$|\.env$" -SimpleMatch:$false
if ($leaks) {
  Write-Host "ERROR: sensitive files detected in staging area:"
  $leaks | ForEach-Object { Write-Host "  $_" }
  Write-Host "Aborting. Check your .gitignore."
  exit 1
}
Write-Host "[OK] No private keys or secrets detected"

# 5. Commit
Write-Host "[..] commit"
$commitMsg = "feat: initial Hephaestus PoC`n`nPermissioned-blockchain proof-of-concept that re-architects Bank Indonesia Antasena regulatory reporting on Hyperledger Fabric. Includes Go chaincode, Docker Compose network (3 orderers + 5 bank peers + OJK observer), Node.js REST gateway, web dashboard, Ansible playbook for multi-VM RHEL deploys, bilingual docs (ID+EN), 13-slide pptx deck, and LinkedIn launch copy. Apache 2.0. Independent PoC, not affiliated with BI or OJK."
git commit -m $commitMsg | Out-Null
$sha = git rev-parse --short HEAD
Write-Host "[OK] commit $sha"

# 6. Remote
Write-Host "[..] adding remote origin"
$existing = git remote 2>$null
if ($existing -contains "origin") {
  $null = (git remote remove origin 2>&1)
}
git remote add origin $RepoUrl
Write-Host "[OK] remote origin -> $RepoUrl"

# 7. Push
Write-Host ""
Write-Host "[..] git push -u origin $Branch"
Write-Host "    A browser popup will appear for GitHub authentication."
Write-Host "    Login with your GitHub account once and Windows will remember it."
Write-Host ""

git push -u origin $Branch

if ($LASTEXITCODE -eq 0) {
  $webUrl = $RepoUrl -replace "^git@github\.com:", "https://github.com/" -replace "\.git$", ""
  Write-Host ""
  Write-Host "======================================================="
  Write-Host " SUCCESS - your repo is live on GitHub"
  Write-Host "======================================================="
  Write-Host ""
  Write-Host "  Open in browser: $webUrl"
  Write-Host ""
  Write-Host "  Suggested next steps:"
  Write-Host "    1. At GitHub.com, edit the repo description and add topics:"
  Write-Host "       blockchain, hyperledger-fabric, fintech, regtech,"
  Write-Host "       banking, indonesia, antasena, bank-indonesia"
  Write-Host "    2. Tag the first release:"
  Write-Host "         git tag -a v0.1.0-poc -m 'Initial PoC release'"
  Write-Host "         git push origin v0.1.0-poc"
  Write-Host "    3. Create a GitHub Release from the tag and attach"
  Write-Host "       pptx/hephaestus-deck.pptx as a downloadable asset."
  Write-Host ""
} else {
  Write-Host ""
  Write-Host "ERROR: push failed. Check the message above."
  Write-Host ""
  Write-Host "Common causes:"
  Write-Host "  - Wrong repo URL (verify at github.com/your-username/repo-name)"
  Write-Host "  - Not logged in (the credential manager popup should have appeared)"
  Write-Host "  - Repo was not empty (create a NEW repo without README/license/.gitignore)"
  exit 1
}
