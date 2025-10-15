# Maintainer: Taufik Hidayat <tfkhdyt@proton.me>

pkgname=opencommit-bin
pkgver=0.1.0
pkgrel=1
pkgdesc='A CLI that writes your git commit messages for you with OpenAI-compatible AI providers'
arch=('x86_64' 'aarch64')
url='https://github.com/lorne-luo/opencommit'
license=('GPL3')
depends=('git')
source_x86_64=("${pkgname}-v${pkgver}.tar.gz::${url}/releases/download/v${pkgver}/opencommit-v${pkgver}-linux-amd64.tar.gz")
sha256sums_x86_64=('2a467b4a5b3d56f76a50ee61fa964832b8326912ec3854f87e4b3acc16cd3089')

source_aarch64=("${pkgname}-v${pkgver}.tar.gz::${url}/releases/download/v${pkgver}/opencommit-v${pkgver}-linux-arm64.tar.gz")
sha256sums_aarch64=('26bb27cb663e1fed462fe277b7c4965077e68d62845f138a9fa9a751bf48dbf7')

build() {
	./opencommit completion bash >opencommit.bash
	./opencommit completion zsh >_opencommit.zsh
	./opencommit completion fish >opencommit.fish
}

package() {
	install -Dm755 opencommit "${pkgdir}/usr/bin/opencommit"
	install -Dm644 opencommit.bash "${pkgdir}/usr/share/bash-completion/completions/opencommit"
	install -Dm644 _opencommit.zsh "${pkgdir}/usr/share/zsh/site-functions/_opencommit"
	install -Dm644 opencommit.fish "${pkgdir}/usr/share/fish/vendor_completions.d/opencommit.fish"
}
