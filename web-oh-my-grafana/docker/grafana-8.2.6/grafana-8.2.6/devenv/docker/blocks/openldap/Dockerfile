# Fork of https://github.com/dinkel/docker-openldap

FROM debian:jessie

LABEL maintainer="Grafana team <hello@grafana.com>"

ENV OPENLDAP_VERSION 2.4.40

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -y \
        slapd=${OPENLDAP_VERSION}* \
        ldap-utils && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN mv /etc/ldap /etc/ldap.dist

EXPOSE 389

VOLUME ["/etc/ldap", "/var/lib/ldap"]

COPY ldap.conf /etc/ldap.dist/ldap.conf

COPY modules/ /etc/ldap.dist/modules
COPY prepopulate/ /etc/ldap.dist/prepopulate

COPY entrypoint.sh /entrypoint.sh
COPY prepopulate.sh /prepopulate.sh

ENTRYPOINT ["/entrypoint.sh"]

CMD ["slapd", "-d", "32768", "-u", "openldap", "-g", "openldap"]
