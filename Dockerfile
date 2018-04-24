FROM alpine:3.6

RUN	apk update && \
	apk --no-cache --update add \
		ca-certificates \
		curl \
		git \
		openssl \
		openssh-client \
		p7zip \
		python \
		py-lxml \
		py-pip \
		rsync \
		sshpass \
		vim \
		zip \		
    && apk --update add --virtual \
		build-dependencies \
		python-dev \
		libffi-dev \
		openssl-dev \
		build-base \
		autoconf \
		automake \
	&& pip install --upgrade \
		pip \
		cffi \
		botocore \
	&& pip install \
		ansible==2.5.0 \
		ansible-lint==3.4.17 \
		awscli==1.15.4 \
		boto==2.48.0 \
		boto3==1.7.4 \
		docker==2.7.0 \
		dopy==0.3.7 \
    && mkdir -p /tmp/download \
    && curl -L https://download.docker.com/linux/static/stable/x86_64/docker-18.03.0-ce.tgz | tar -xz -C /tmp/download \
    && mv /tmp/download/docker/docker /usr/local/bin/ \
    && cd /tmp/download \
	&& git clone https://github.com/bryanpkc/corkscrew.git \
	&& cd corkscrew \
	&& autoreconf --install && ./configure && make install \
	&& apk del build-dependencies \
	&& rm -rf /tmp/* \
	&& rm -rf /var/cache/apk/*


RUN mkdir -p /opt/lagoon/bin

COPY ./go/installer /opt/lagoon/bin/installer
COPY ./ansible/ /opt/lagoon/ansible

ENTRYPOINT ["/opt/lagoon/bin/installer"]


