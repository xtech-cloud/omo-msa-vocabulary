package config

const defaultJson string = `{
	"service": {
		"address": ":7073",
		"ttl": 15,
		"interval": 10
	},
	"logger": {
		"level": "info",
		"file": "logs/server.log",
		"std": false
	},
	"database": {
		"name": "rgsCloud",
		"ip": "127.0.0.1",
		"port": "27017",
		"user": "root",
		"password": "pass2019",
		"type": "mongodb"
	},
	"graph": {
		"name": "RedGraph",
		"user": "neo4j",
		"password": "yumei2020",
		"ip": "127.0.0.1",
		"port": "11005"
	},
	"basic": {
		"tags": 6,
		"synonyms": 5
	}
}
`
