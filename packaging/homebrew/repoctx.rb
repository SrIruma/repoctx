class Repoctx < Formula
  desc "Keep your AI coding-agent context files (AGENTS.md, CLAUDE.md) truthful"
  homepage "https://github.com/SrIruma/repoctx"
  version "0.8.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/SrIruma/repoctx/releases/download/v0.8.0/repoctx_darwin_arm64"
      sha256 "6101ebca9a70005a5eab44141c6654f4427ec9ed1465ea24092d17c0225247ea"
    else
      url "https://github.com/SrIruma/repoctx/releases/download/v0.8.0/repoctx_darwin_amd64"
      sha256 "798530de57b90bd847deaadbde22795c83569b1cd6bfc832698f88564130a4c3"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/SrIruma/repoctx/releases/download/v0.8.0/repoctx_linux_arm64"
      sha256 "e2b64cdc6d66ee406a436a3623229dd511e3ed4ac97fd8f5f783654eff0f711b"
    else
      url "https://github.com/SrIruma/repoctx/releases/download/v0.8.0/repoctx_linux_amd64"
      sha256 "51d22c6ba0309b0e3b21a453468632616e809d598bfa68dac4a7c0216550a597"
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
