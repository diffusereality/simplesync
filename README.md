# SimpleSync

This is a simple utility that syncs k8s manifests. It doesnt provide any utility on top of it like Argo of Flux, it just gets from github, compares versions and if they differ, runs kubectl apply -f file.yml