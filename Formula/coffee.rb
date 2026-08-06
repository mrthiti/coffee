class Coffee < Formula
  desc "Keep your Mac awake even with the lid closed"
  homepage "https://github.com/mrthiti/coffee"
  url "https://github.com/mrthiti/coffee/archive/897389b98d09ea4d0196d4a5a173584538ccfa21.tar.gz"
  version "0.1.0"
  sha256 "1f9ad7eb7a4750695d13018c820c510d299b0073d24ddadb4b4ea54d86dd3f99"
  license "MIT"

  depends_on "go" => :build
  depends_on :macos

  def install
    system "go", "build", *std_go_args(ldflags: "-X main.version=v#{version}"), "."
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/coffee --version")
  end
end
