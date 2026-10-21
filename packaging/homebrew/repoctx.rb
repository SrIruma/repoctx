class Repoctx < Formula
  desc "Keep your AI coding-agent context files (AGENTS.md, CLAUDE.md) truthful"
  homepage "https://github.com/SrIruma/repoctx"
  version "0.7.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/SrIruma/repoctx/releases/download/v0.7.0/repoctx_darwin_arm64"
      sha256 "b36bebdb35a36604e85e897ddddfb541da4ec751c06e75edfb4538b7e437e89b"
    else
      url "https://github.com/SrIruma/repoctx/releases/download/v0.7.0/repoctx_darwin_amd64"
      sha256 "9b931983e4e0b9589c646583dc9407257b367ee8d59d6c0922445806f18ce27c"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/SrIruma/repoctx/releases/download/v0.7.0/repoctx_linux_arm64"
      sha256 "8155520de2b1a95ca957088a75cc532174053449ad8fe64c76942adafdc55e02"
    else
      url "https://github.com/SrIruma/repoctx/releases/download/v0.7.0/repoctx_linux_amd64"
      sha256 "0b7f2d0ff066b759d2c4c85cf7f8bec6b101326b7f9757cebc02299f657e3637"
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
