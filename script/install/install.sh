#!/usr/bin/env bash
set -euo pipefail

if [[ ${OS:-} = Windows_NT ]]; then
    echo 'error: please install xgo using Windows Subsystem for Linux'
    exit 1
fi

error() {
    echo "$@" >&2
    exit 1
}

command -v tar >/dev/null || error 'tar is required to install xgo'

case $(uname -ms) in
'Darwin x86_64')
    target=darwin-amd64
    ;;
'Darwin arm64')
    target=darwin-arm64
    ;;
'Linux aarch64' | 'Linux arm64')
    target=linux-arm64
    ;;
'Linux x86_64' | *)
    target=linux-amd64
    ;;
esac

latestURL="https://github.com/xhd2015/xgo/releases/latest"
headers=$(curl "$latestURL" -so /dev/null -D -)
if [[ "$headers" != *302* ]];then
    error "expect 302 from $latestURL"
fi

location=$(echo "$headers"|grep "location: ")
if [[ -z $location ]];then
    error "expect 302 location from $latestURL"
fi
locationURL=${location/#"location: "}
locationURL=${locationURL/%$'\n'}
locationURL=${locationURL/%$'\r'}

versionName=""
if [[ "$locationURL" = *'/xgo-v'* ]];then
    versionName=${locationURL/#*'/xgo-v'}
elif [[ "$locationURL" = *'/tag/v'* ]];then
    versionName=${locationURL/#*'/tag/v'}
fi

if [[ -z $versionName ]];then
   error "expect tag format: xgo-v1.x.x, actual: $locationURL"
fi

file="xgo${versionName}-${target}.tar.gz"
uri="$latestURL/download/$file"
install_dir=$HOME/.xgo
bin_dir=$install_dir/bin

if [[ ! -d $bin_dir ]]; then
    mkdir -p "$bin_dir" || error "failed to create install directory \"$bin_dir\""
fi

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

curl --fail --location --progress-bar --output "${tmp_dir}/${file}" "$uri" || error "failed to download xgo from \"$uri\""

(
    cd "$bin_dir"
    tar -xzf "${tmp_dir}/${file}"
    if [[ -f xgo ]];then chmod +x ./xgo;fi
)

if [[ "$INSTALL_TO_BIN" == "true" ]];then
    # install fails if target already exists
    if [[ -f /usr/local/bin/xgo ]];then
        mv /usr/local/bin/{xgo,xgo_backup}
    fi
    install "$bin_dir/xgo" /usr/local/bin
else
    if [[ -f ~/.bash_profile ]];then
        content=$(cat ~/.bash_profile)
        if [[ "$content" != *'# setup xgo'* ]];then
            echo "# setup xgo" >> ~/.bash_profile
            echo "export PATH=\"$bin_dir:\$PATH\"" >> ~/.bash_profile
        fi
    fi

    if [[ -f ~/.zshrc ]];then
        content=$(cat ~/.zshrc)
        if [[ "$content" != *'# setup xgo'* ]];then
            echo "# setup xgo" >> ~/.zshrc
            echo "export PATH=\"$bin_dir:\$PATH\"" >> ~/.zshrc
        fi
    fi
fi


xgoTip=xgo
sourceTip=""
if ! command -v xgo >/dev/null;then
     xgoTip="~/.xgo/bin/xgo"
     sourceTip=$'You may need to source shell profile or add xgo to PATH to use:\n    export PATH=~/.xgo/bin:$PATH'
fi

echo "Successfully installed, to get started, run:"
echo "  $xgoTip help"
if [[ -n "$sourceTip" ]];then
    echo ""
    echo "NOTE: $sourceTip"
fi
echo ""