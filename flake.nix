{
  description = "Event Sensor - concert tracker, Go + SQLite + Vue";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    playwright.url = "github:pietdevries94/playwright-web-flake";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      playwright,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        playwrightDeps = playwright.packages.${system};
      in
      {
        # bun and node are assumed on the host and stay out of this list. The shell
        # provides what the machine does not: the Go toolchain, the SQL tools, and a
        # Playwright whose browsers are dynamically linked against this store.
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.gopls
            pkgs.sqlc
            pkgs.goose
            pkgs.air

            playwrightDeps.playwright-test

            pkgs.just
            pkgs.jq
            pkgs.sqlite
          ];

          shellHook = ''
            export PLAYWRIGHT_BROWSERS_PATH="${playwrightDeps.playwright-driver.browsers}"
            export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
            export PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=true
          '';
        };
      }
    );
}
