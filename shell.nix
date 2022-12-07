with import <nixpkgs> {};
#{ stdenv, pkgs, lib }:

pkgs.mkShell {
  buildInputs = with pkgs; [

    # Nodejs
    nodejs-18_x
    (yarn.override { nodejs = nodejs-18_x; })
    nodePackages.typescript
    nodePackages.typescript-language-server
    nodePackages.npm

    # golang
    go_1_19
    gopls
    delve
    golangci-lint
    gotools

    # k3d
    kube3d

    # kubernetes tools
    kubectl
    k9s

    # helm
    kubernetes-helm
  ];
}
