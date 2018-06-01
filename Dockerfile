FROM tbouvet/ansible-docker:1.0-rc2

RUN mkdir -p /opt/lagoon/bin
COPY ./go/installer /opt/lagoon/bin/installer

RUN mkdir -p /opt/lagoon/ansible
WORKDIR /opt/lagoon/ansible
RUN git clone -b feat-create https://github.com/tbouvet/aws-provider.git 
RUN git clone -b feat-create https://github.com/tbouvet/core.git 
RUN chmod -R 755 */scripts

ENTRYPOINT ["/opt/lagoon/bin/installer"]


