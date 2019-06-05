FROM centos:centos7

RUN yum -y install \
  epel-release \
  && yum clean all && \
  yum -y install  \
  git \
  man-db \
  curl \
  unzip \
  tar \
  vim-enhanced \
  sudo

CMD ["/bin/bash"]
