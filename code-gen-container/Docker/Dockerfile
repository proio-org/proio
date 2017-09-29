FROM dbcooper/arch:2017-09-08

# Set up basic environment
RUN pacman -S --noconfirm \
		sed \
		sudo

RUN useradd -m -G wheel builduser && \
	sed -i.bak 's/# \(%wheel ALL=(ALL) NOPASSWD: ALL\)/\1/' /etc/sudoers

USER builduser
WORKDIR /home/builduser

CMD /bin/bash

# Get build tools

RUN sudo pacman -S --noconfirm \
		clang-tools-extra \
		git \
		go \
		grep \
		make \
		protobuf && \
	sudo rm -rf /var/cache/pacman/pkg/*

ENV GOPATH /opt/Go
ENV PATH "$GOPATH/bin:${PATH}"

RUN go get github.com/gogo/protobuf/protoc-gen-gofast && \
	rm -rf $GOPATH/src
