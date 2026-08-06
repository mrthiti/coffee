cask "coffee" do
  version "1.0.0"

  on_arm do
    url "https://github.com/mrthiti/coffee/releases/download/v#{version}/coffee-darwin-arm64"
    sha256 "b0c96815c3bc2f957839e6f3488714df92eb6238ba059bada5670007f49ff9c8"
    binary "coffee-darwin-arm64", target: "coffee"
  end

  on_intel do
    url "https://github.com/mrthiti/coffee/releases/download/v#{version}/coffee-darwin-amd64"
    sha256 "e17a3301802be589b24bd439059cd2719d85cff5eb1f6aeef0a78c9b2b22f84f"
    binary "coffee-darwin-amd64", target: "coffee"
  end

  name "coffee"
  desc "Keep your Mac awake even with the lid closed"
  homepage "https://github.com/mrthiti/coffee"

  postflight do
    user = ENV["USER"] || Utils.safe_popen_read("id", "-un").strip
    rule = "#{user} ALL=(root) NOPASSWD: /usr/bin/pmset -a disablesleep 0, /usr/bin/pmset -a disablesleep 1\n"

    Dir.mktmpdir "coffee-sudoers" do |dir|
      rule_file = "#{dir}/coffee"
      File.write(rule_file, rule)
      File.chmod(0600, rule_file)

      if system_command("/usr/sbin/visudo", args: ["-c", "-f", rule_file], print_stderr: false).success?
        ohai "Granting passwordless sudo for `pmset -a disablesleep` (you may be asked for your password)"
        system_command "/usr/bin/install", args: ["-m", "440", rule_file, "/etc/sudoers.d/coffee"], sudo: true
      else
        opoo "Generated sudoers rule failed validation — skipping passwordless sudo setup. Run `sudo coffee` instead."
      end
    end
  end

  uninstall_postflight do
    system_command "/bin/rm", args: ["-f", "/etc/sudoers.d/coffee"], sudo: true
  end

  caveats do
    <<~EOS
      coffee needs to run `pmset -a disablesleep` as root. This cask tried to
      grant your user passwordless sudo for exactly that command during install
      (you may have been prompted for your password).

      If that step was skipped or declined, plain `coffee` will only work when
      you already have a cached sudo credential — otherwise run `sudo coffee`
      instead, or add the sudoers rule yourself (see README).
    EOS
  end
end
