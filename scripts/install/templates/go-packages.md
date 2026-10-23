| System | Install Go and git |
| --- | --- |
| Debian 13 and systems based on it | `sudo apt-get update && sudo apt-get install -y golang-go git ca-certificates` |
| Fedora 44 | `sudo dnf install -y golang git` |
| Arch Linux (official repositories, no AUR) | `sudo pacman -S --needed go git` |
| Gentoo | `sudo emerge --ask --noreplace dev-lang/go dev-vcs/git` |
| Void Linux | `sudo xbps-install -Sy go git` |
| openSUSE Tumbleweed and Leap 16 | `sudo zypper install -y go git` |
| Alpine Linux | `doas apk add go git` |
| NixOS | none: the NixOS command below brings Go and git |
| Guix System | none: the Guix command below brings Go and git |
| macOS | `brew install go git` |
| FreeBSD | `su -m root -c 'pkg install -y go git'` |
| OpenBSD | `doas pkg_add go git` |
| NetBSD | `su -m root -c 'pkgin -y install go git' && PATH=$(ls -d /usr/pkg/go1*/bin \| tail -n 1):$PATH && export PATH` |
| Windows (Command Prompt or PowerShell 7) | `winget install --id GoLang.Go -e && winget install --id Git.Git -e` |
| Windows PowerShell 5.1 | `winget install --id GoLang.Go -e; if ($?) { winget install --id Git.Git -e }` |
