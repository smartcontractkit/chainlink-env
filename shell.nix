{ stdenv, pkgs, lib }:

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

    # aws cli
    awscli2

    # k3d
    kube3d

    # kubernetes tools
    kubectl
    k9s

    # helm
    kubernetes-helm
  ];

  shellHook = ''
    # Setup helm repositories
    helm repo add chainlink-qa https://raw.githubusercontent.com/smartcontractkit/qa-charts/gh-pages/
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm repo update
  '';
}
