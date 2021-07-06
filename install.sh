#!/bin/sh
# This script is inspired on the Deno installer: (https://deno.land/x/install@v0.1.4/install.sh)

set -e

case $(uname -sm) in
"Darwin x86_64") target="darwin_amd64" ;;
"Darwin arm64") target="darwin_arm64" ;;
*) target="linux_amd64" ;;
esac


version=$(curl https://api.github.com/repos/koyeb/koyeb-cli/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | cut -c 2-)
koyeb_url="https://github.com/koyeb/koyeb-cli/releases/latest/download/koyeb-cli_${version}_${target}.tar.gz"

koyeb_install="${KOYEB_INSTALL:-$HOME/.koyeb}"
bin_dir="$koyeb_install/bin"
exe="$bin_dir/koyeb"

if [ ! -d "$bin_dir" ]; then
	mkdir -p "$bin_dir"
fi

curl --fail --location --progress-bar --output "$exe.tar.gz" "$koyeb_url"
tar xvf "$exe.tar.gz" -C "$bin_dir"
chmod +x "$exe"
rm "$exe.tar.gz"

echo "Koyeb CLI was installed successfully to $exe"
if command -v koyeb >/dev/null; then
	echo "Run 'koyeb --help' to get started"
else
	case $SHELL in
	/bin/zsh) shell_profile=".zshrc" ;;
	*) shell_profile=".bash_profile" ;;
	esac
	echo "Manually add the directory to your \$HOME/$shell_profile (or similar)"
	echo "  export PATH=\"$bin_dir:\$PATH\""
	echo "Run '$exe --help' to get started"
fi

