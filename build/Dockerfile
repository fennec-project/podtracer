FROM centos:8

RUN yum install -y tcpdump

COPY podtracer /usr/bin/podtracer
COPY kubeconfig ~/.kube/kubeconfig

USER root