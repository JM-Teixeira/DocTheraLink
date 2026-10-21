@echo off
rem Starts the local Klyniva cluster inside WSL and opens it in the browser.
rem Keep this window open while you use the app: WSL (and with it Docker and
rem the cluster) shuts down when no window is attached to it.
rem Override with: set KLYNIVA_DISTRO=... / set KLYNIVA_REPO=...
if "%KLYNIVA_DISTRO%"=="" set KLYNIVA_DISTRO=Ubuntu-22.04
if "%KLYNIVA_REPO%"=="" set KLYNIVA_REPO=/root/docthera/DocTheraLink
title Klyniva (keep open)
wsl -d %KLYNIVA_DISTRO% -u root --cd %KLYNIVA_REPO% -- bash scripts/k8s-local.sh start
if errorlevel 1 (
  echo.
  echo Klyniva did not start. If the cluster was never created run in WSL:  make k8s-up
  pause
  exit /b 1
)
start "" http://localhost:8088
echo Keep this window open while using Klyniva. Close it (or press Ctrl+C) when done.
wsl -d %KLYNIVA_DISTRO% -u root -- sleep infinity
