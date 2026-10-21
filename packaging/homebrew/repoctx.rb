class Repoctx < Formula
  desc "Keep your AI coding-agent context files (AGENTS.md, CLAUDE.md) truthful"
  homepage "https://github.com/SrIruma/repoctx"
  version "0.9.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/SrIruma/repoctx/releases/download/v0.9.0/repoctx_darwin_arm64"
      sha256 "0b985a5789b7db7fc28a1dc1ce55352f72765c1040436bf099e87fad463a3c27"
    else
      url "https://github.com/SrIruma/repoctx/releases/download/v0.9.0/repoctx_darwin_amd64"
      sha256 "dbc118695496b2e3e94ee06536e704c7406582e22d80cc3eabcc5909b1b032e9"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/SrIruma/repoctx/releases/download/v0.9.0/repoctx_linux_arm64"
      sha256 "9367d3de5e15224aaa1f748ada46edcdf48296dd41b3e032b237862035b4e993"
    else
      url "https://github.com/SrIruma/repoctx/releases/download/v0.9.0/repoctx_linux_amd64"
      sha256 "6e16e49d11ff49113a0d774e8208c55f95b7aeaa37ba90813ccd9514eb8810e8"
    end
  end

  def install
    binary = Dir["repoctx_*"].first
    bin.install binary => "repoctx"
  end

  test do
    assert_match "repoctx", shell_output("#{bin}/repoctx --version")
  end
end
