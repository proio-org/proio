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
		cmake \
		git \
		go \
		grep \
		make && \
	sudo rm -rf /var/cache/pacman/pkg/*

ENV GOPATH /opt/Go
ENV PATH "$GOPATH/bin:${PATH}"

RUN go get github.com/gogo/protobuf/protoc-gen-gofast && \
	rm -rf $GOPATH/src

ENV PROTOBUF_RELEASE v3.4.1

RUN sudo pacman -S --noconfirm \
		autoconf \
		automake \
		unzip && \
	sudo rm -rf /var/cache/pacman/pkg/*

RUN git clone https://github.com/google/protobuf.git && \
	cd protobuf && \
	git checkout tags/$PROTOBUF_RELEASE && \
	./autogen.sh && \
	./configure && \
	make -j4 && \
	sudo make install && \
	cd .. && \
	rm -rf protobuf
