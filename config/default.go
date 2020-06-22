package config

const defaultYAML string = `
service:
    address: :7079
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
    name: RedGraph
    user: neo4j
    password: "12345678"
    ip: localhost
    port: "11003"
cache:
    kind: 1
    domain: "http://testdown.suii.cn"
    bucket: tec-test
    accessKey: 4TDqfvaNHKxzx4nFz0YglS_jHlKXECCSSWb1vUr5
    secretKey: pZ8AnJE5IYgNRUFEB132ohIToJdRe5uxm4ZLLljp
basic:
    synonym: 6
    tag: 6
    
`
