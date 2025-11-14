ARG OMNI_VERSION=dev

FROM debian:bookworm-slim

ARG OMNI_VERSION
ENV OMNI_HOME=/opt/omni
WORKDIR ${OMNI_HOME}

COPY dist/omni-${OMNI_VERSION}/ ${OMNI_HOME}/

RUN ln -s ${OMNI_HOME}/bin/omnic /usr/local/bin/omnic && \
    ln -s ${OMNI_HOME}/bin/omnir /usr/local/bin/omnir && \
    ln -s ${OMNI_HOME}/bin/omnipkg /usr/local/bin/omnipkg

ENV PATH="${OMNI_HOME}/bin:${PATH}"

RUN groupadd --system omni && \
    useradd --system --gid omni --home-dir ${OMNI_HOME} --shell /bin/bash omni && \
    chown -R omni:omni ${OMNI_HOME}

USER omni

CMD ["omnic", "--help"]
