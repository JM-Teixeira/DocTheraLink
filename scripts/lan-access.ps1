<#
.SYNOPSIS
  Lets other computers on the same (private) network open the local Klyniva.

.DESCRIPTION
  Klyniva runs in a k3d cluster inside WSL2 and answers on http://localhost:8088
  of this PC only. This script (run as Administrator):
    1. forwards port 8088 of this PC's network cards into WSL (netsh portproxy),
       unless WSL already runs in "mirrored" networking mode;
    2. allows TCP 8088 in Windows Firewall on PRIVATE networks only (never on
       public Wi-Fi);
    3. tells the backend its address on the network (links in e-mails, CORS);
    4. prints the addresses to open from another laptop.
  The WSL address changes after a reboot: run it again then (it is safe to
  repeat). -Remove undoes everything.

  Demo data and the demo password are for a home network only. Before real
  patient data: HTTPS and demo users off (see TODO.md).

.EXAMPLE
  powershell -ExecutionPolicy Bypass -File scripts\lan-access.ps1
.EXAMPLE
  powershell -ExecutionPolicy Bypass -File scripts\lan-access.ps1 -Remove
#>
param(
  [int]$Port = 8088,
  [string]$Distro = "Ubuntu-22.04",
  [string]$Repo = "/root/docthera/DocTheraLink",
  [switch]$Remove
)

$ErrorActionPreference = "Stop"
$ruleName = "Klyniva (TCP $Port, private networks)"

$admin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole(
  [Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $admin) {
  Write-Host "Run this from an Administrator PowerShell (it changes the firewall and port forwarding)." -ForegroundColor Yellow
  exit 1
}

if ($Remove) {
  netsh interface portproxy delete v4tov4 listenport=$Port listenaddress=0.0.0.0 2>$null | Out-Null
  Get-NetFirewallRule -DisplayName $ruleName -ErrorAction SilentlyContinue | Remove-NetFirewallRule
  wsl -d $Distro -u root --cd $Repo -- bash scripts/k8s-local.sh lan off
  Write-Host "Klyniva is again reachable from this computer only (http://localhost:$Port)."
  exit 0
}

# 1. Port forwarding into WSL (not needed in mirrored mode)
$wslconfig = Join-Path $env:USERPROFILE ".wslconfig"
$mirrored = (Test-Path $wslconfig) -and ((Get-Content $wslconfig -Raw) -match "networkingMode\s*=\s*mirrored")
if ($mirrored) {
  Write-Host "WSL uses mirrored networking: no port forwarding needed."
} else {
  $wslIp = (wsl -d $Distro -u root -- hostname -I).Trim().Split(" ")[0]
  if (-not $wslIp) { throw "Could not read the WSL address. Is WSL running?" }
  netsh interface portproxy delete v4tov4 listenport=$Port listenaddress=0.0.0.0 2>$null | Out-Null
  netsh interface portproxy add v4tov4 listenport=$Port listenaddress=0.0.0.0 connectport=$Port connectaddress=$wslIp | Out-Null
  Write-Host "Forwarding port $Port to WSL ($wslIp)."
}

# 2. Firewall: private networks only
if (-not (Get-NetFirewallRule -DisplayName $ruleName -ErrorAction SilentlyContinue)) {
  New-NetFirewallRule -DisplayName $ruleName -Direction Inbound -Protocol TCP -LocalPort $Port `
    -Action Allow -Profile Private | Out-Null
}
$profiles = Get-NetConnectionProfile | Where-Object { $_.IPv4Connectivity -ne "Disconnected" }
if ($profiles | Where-Object { $_.NetworkCategory -eq "Public" }) {
  Write-Host "Note: this network is marked Public in Windows, so other computers are blocked." -ForegroundColor Yellow
  Write-Host "      If it is your home/clinic network: Settings > Network > Properties > Private network."
}

# 3. Tell the backend its address on the network
$name = "$($env:COMPUTERNAME.ToLower()).local"
# Real network cards only (not loopback, WSL/Hyper-V/Docker virtual ones)
$ips = Get-NetIPAddress -AddressFamily IPv4 |
  Where-Object { $_.InterfaceAlias -notmatch "Loopback|vEthernet|WSL|Docker" -and $_.IPAddress -notmatch "^169\.254\." } |
  Select-Object -ExpandProperty IPAddress
$hosts = @($name) + $ips
wsl -d $Distro -u root --cd $Repo -- bash scripts/k8s-local.sh lan ($hosts -join ",")

# 4. Where to go
Write-Host ""
Write-Host "Open Klyniva from another computer on this network:" -ForegroundColor Green
Write-Host "  http://${name}:$Port"
foreach ($ip in $ips) { Write-Host "  http://${ip}:$Port" }
Write-Host ""
Write-Host "If the .local name does not open, use the number. After a reboot run this script again."
