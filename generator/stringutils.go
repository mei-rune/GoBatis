package generator

import (
	"strings"

	"github.com/grsmv/inflect"
)

func CamelCase(name string) string {
	switch name {
	case "db2":
		return "DB2"
	case "iis":
		return "IIS"
	case "tcp":
		return "TCP"
	case "http":
		return "HTTP"
	case "ftp":
		return "FTP"
	case "smtp":
		return "SMTP"
	case "pop3":
		return "POP3"
	case "imap":
		return "IMAP"
	case "dhcp":
		return "DHCP"
	case "dns":
		return "DNS"
	case "ldap":
		return "LDAP"
	case "epon":
		return "EPON"
	case "mpls_pe":
		return "MplsPE"
	case "mpls_ce":
		return "MplsCE"
	}
	return inflect.Camelize(name)
}

func Underscore(name string) string {
	switch name {
	case "DB2":
		return "db2"
	case "IIS":
		return "iis"
	case "TCP":
		return "tcp"
	case "HTTP":
		return "http"
	case "FTP":
		return "ftp"
	case "SMTP":
		return "smtp"
	case "POP3":
		return "pop3"
	case "IMAP":
		return "imap"
	case "DHCP":
		return "dhcp"
	case "DNS":
		return "dns"
	case "LDAP":
		return "ldap"
	case "EPON":
		return "epon"
	case "MplsPE":
		return "mpls_pe"
	case "MplsCE":
		return "mpls_ce"
	}
	uname := inflect.Underscore(name)
	return strings.Replace(uname, "_i_d", "_id", -1)
}

func Pluralize(name string) string {
	switch name {
	case "db2":
		return "db2"
	case "windows":
		return "windows"
	case "iis":
		return "iis"
	case "tcp":
		return "tcp"
	case "http":
		return "http"
	case "ftp":
		return "ftp"
	case "smtp":
		return "smtp"
	case "pop3":
		return "pop3"
	case "imap":
		return "imap"
	case "dhcp":
		return "dhcp"
	case "dns":
		return "dns"
	case "ldap":
		return "ldap"
	case "epon":
		return "epon"
	case "mpls_pe":
		return "mpls_pe"
	case "mpls_ce":
		return "mpls_ce"
	}

	return inflect.Pluralize(name)
}

func Tableize(className string) string {
	switch className {
	case "DB2":
		return "db2"
	case "IIS":
		return "iis"
	case "TCP":
		return "tcp"
	case "HTTP":
		return "http"
	case "FTP":
		return "ftp"
	case "SMTP":
		return "smtp"
	case "POP3":
		return "pop3"
	case "IMAP":
		return "imap"
	case "DHCP":
		return "dhcp"
	case "DNS":
		return "dns"
	case "LDAP":
		return "ldap"
	case "EPON":
		return "epon"
	case "MplsPE":
		return "mpls_pe"
	case "MplsCE":
		return "mpls_ce"
	}
	return Pluralize(Underscore(className))
}

func Singularize(word string) string {
	if "windows" == word {
		return "windows"
	}
	if "iis" == word {
		return "iis"
	}
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
