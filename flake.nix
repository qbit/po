{
  description = "po: pushover notification tool";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      supportedSystems =
        [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in {
      overlay = final: prev: {
        po = self.packages.${prev.system}.po;
      };
      packages = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          po = pkgs.buildGoModule {
            pname = "po";
            version = "v1.0.0";
            src = ./.;

            vendorHash = "sha256-ftBMG3c5uDkNiOiZvO/ixa8FOrMfE8VM1AmLTSlYu2o=";
          };
        });

      defaultPackage = forAllSystems (system: self.packages.${system}.po);
      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            shellHook = ''
              PS1='\u@\h:\@; '
              echo "Go `${pkgs.go}/bin/go version`"
            '';
            nativeBuildInputs = with pkgs; [ git go gopls go-tools ];
          };
        });
    };
}

