package config

const defaultYAML string = `
service:
    address: :7073
    ttl: 15
    interval: 10
logger:
    level: info
    dir: /var/log/msa/
database:
    name: rgsCloud
    ip: 127.0.0.1
    port: "27017"
    user: root
    password: pass2019
    type: mongodb
graph:
    name: RdGraph
    user: neo4j
    password: "yumei2020"
    ip: 39.96.113.123
    port: "11003"
basic:
  tags: 6
  synonyms: 5
`
