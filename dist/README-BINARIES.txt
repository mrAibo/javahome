javahome v0.5.1 precompiled binaries

Use a precompiled binary for your OS/CPU, or build from source yourself.

Windows:
  javahome-windows-amd64.exe   Most normal Intel/AMD Windows PCs
  javahome-windows-arm64.exe   Windows on ARM

Linux:
  javahome-linux-amd64         Intel/AMD 64-bit Linux
  javahome-linux-arm64         ARM64 Linux

macOS:
  javahome-darwin-amd64        Intel Mac
  javahome-darwin-arm64        Apple Silicon Mac

Quick Windows example:
  .\javahome-windows-amd64.exe help
  .\javahome-windows-amd64.exe list
  .\javahome-windows-amd64.exe use 17 --shell powershell | Invoke-Expression
  .\javahome-windows-amd64.exe windows-env user 17 --dry-run

Build yourself:
  git clone https://github.com/mrAibo/javahome.git
  cd javahome
  go test ./...
  go build -o bin/javahome ./cmd/javahome

Checksums are in checksums.txt.
