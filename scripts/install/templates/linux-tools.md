| Distribution | Install what the command needs (once) |
| --- | --- |
| Debian 13 and systems based on it | `sudo apt-get update && sudo apt-get install -y curl ca-certificates` |
| Fedora 44 (curl is installed) | `sudo dnf install -y tar gzip` |
| Arch Linux | `sudo pacman -S --needed curl tar gzip` |
| Gentoo | `sudo emerge --ask --noreplace net-misc/curl` |
| Void Linux | `sudo xbps-install -Sy curl` |
| openSUSE Tumbleweed and Leap 16 | `sudo zypper install -y curl tar gzip` |
| Alpine Linux | `doas apk add curl` |
| NixOS | `nix-shell -p curl` (a shell with curl; run the command inside it) |
| Guix System | `guix shell curl tar gzip coreutils` (a shell with them; run the command inside it) |
