FROM dxflrs/garage:v2.3.0 AS garage-bin

FROM alpine:3.23

RUN apk add --no-cache bash coreutils curl jq

COPY --from=garage-bin /garage /garage

COPY init.sh /init.sh
RUN chmod +x /init.sh

ENTRYPOINT ["/bin/bash", "/init.sh"]