FROM alpine:3.11
ADD omo.msa.vocabulary /usr/bin/omo.msa.vocabulary
ENV MSA_REGISTRY_PLUGIN
ENV MSA_REGISTRY_ADDRESS
ENTRYPOINT [ "omo.msa.vocabulary" ]
