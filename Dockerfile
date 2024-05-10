FROM btwiuse/arch:golang AS builder-golang

COPY . /wsport

WORKDIR /wsport/cmd

ENV GONOSUMDB="*"

RUN go mod tidy

RUN CGO_ENABLED=0 GOBIN=/usr/local/bin go install -v ./peerid
RUN CGO_ENABLED=0 GOBIN=/usr/local/bin go install -v ./bootstrap

FROM btwiuse/arch

COPY --from=builder-golang /usr/local/bin/peerid /usr/bin/
COPY --from=builder-golang /usr/local/bin/bootstrap /usr/bin/

