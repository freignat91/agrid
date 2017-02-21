FROM alpine:3.4

ENV GOPATH /go
ENV PATH $PATH:/go/bin

RUN echo "@community http://nl.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories

RUN apk update && apk upgrade && \
    apk -v add git make bash go@community musl-dev curl && \
    go version && \
    go get -u github.com/Masterminds/glide/...

COPY ./ /go/src/github.com/freignat91/agrid

RUN cd $GOPATH/src/github.com/freignat91/agrid && \
    rm -f glide.lock && \
    glide install && \
    make install && \
    echo agrid built && \
    chmod +x $GOPATH/bin/* && \
    cd $GOPATH && \
    rm -rf $GOPATH/src && \
    rm -rf /root/.glide

#HEALTHCHECK --interval=10s --timeout=10s --retries=80 CMD /go/bin/server healthcheck

CMD ["/go/bin/server"]
 