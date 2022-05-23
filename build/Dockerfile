FROM fedora
RUN dnf -y clean all
RUN dnf -y update 
RUN dnf -y upgrade 
RUN dnf -y install go nodejs procps supervisor iproute bind-utils nmap net-tools

RUN mkdir -vp /etc/supervisor.d
COPY build/supervisord.conf /etc/
COPY build/*.ini /etc/supervisor.d/
COPY . /src
RUN rm -rf /src/backend/repos/*
RUN echo "gitlab.cee.redhat.com	10.0.128.118" >> /etc/hosts
RUN echo "gitlab.cee.prod.upshift.rdu2.redhat.com	10.0.128.118" >> /etc/hosts
RUN git config --global http.sslVerify false
EXPOSE 1323
EXPOSE 8080
CMD ["supervisord", "-n", "-c", "/etc/supervisord.conf"]
