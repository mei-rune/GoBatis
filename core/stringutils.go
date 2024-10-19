package core

import (
	"strings"

	"github.com/grsmv/inflect"
)

var camelCaseMap = map[string]string{
	"db2":  "DB2",
	"iis":  "IIS",
	"tcp":  "TCP",
	"http": "HTTP",
	"ftp":  "FTP",
	"smtp": "SMTP",
	"pop3": "POP3",
	"imap": "IMAP",
	"dhcp": "DHCP",
	"dns":  "DNS",
	"ldap": "LDAP",
}

var underscores = map[string]string{}

func init() {
	for k, v := range camelCaseMap {
		underscores[v] = k
	}

	underscores["ID"] = "id"
}

func CamelCase(name string) string {
	if camelize, ok := camelCaseMap[name]; ok {
		return camelize
	}
	return inflect.Camelize(name)
}

func Underscore(name string) string {
	if camelize, ok := underscores[name]; ok {
		return camelize
	}

	uname := inflect.Underscore(name)
	return strings.Replace(uname, "_i_d", "_id", -1)
}

func Pluralize(name string) string {
	return inflect.Pluralize(name)
}

func Tableize(className string) string {
	return Pluralize(Underscore(className))
}

func Singularize(word string) string {
	return inflect.Singularize(word)
}

func Capitalize(word string) string {
	return inflect.Capitalize(word)
}

func Typeify(word string) string {
	return inflect.Typeify(word)
}

func CamelizeDownFirst(word string) string {
	return inflect.CamelizeDownFirst(word)
}
