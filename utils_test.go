package gobatis

import (
	"testing"
)

var toGoNamesTestDatas = [][]string{
	{"_", "", ""},
	{"_ID_", "Id", "ID"},
	{"Card_ID_", "CardId", "CardID"},
	{"_ID_name_", "IdName", "IDName"},
	{"id", "Id", "ID"},
	{"hahah_url_id_aaaaa_xss_bb", "HahahUrlIdAaaaaXssBb", "HahahURLIDAaaaaXSSBb"},
	{"foo_bar", "FooBar", "FooBar"},
	{"foo_bar_baz", "FooBarBaz", "FooBarBaz"},
	{"Foo_bar", "FooBar", "FooBar"},
	{"foo_WiFi", "FooWifi", "FooWifi"},
	{"Id", "Id", "ID"},
	{"foo_id", "FooId", "FooID"},
	{"fooId", "Fooid", "Fooid"},
	{"_Leading", "Leading", "Leading"},
	{"___Leading", "Leading", "Leading"},
	{"trailing_", "Trailing", "Trailing"},
	{"trailing___", "Trailing", "Trailing"},
	{"a_b", "AB", "AB"},
	{"a__b", "AB", "AB"},
	{"a___b", "AB", "AB"},
	{"Rpc1150", "Rpc1150", "Rpc1150"},
	{"case3_1", "Case31", "Case31"},
	{"case3__1", "Case31", "Case31"},
	{"IEEE802_16bit", "Ieee80216bit", "Ieee80216bit"},
	{"IEEE802_16Bit", "Ieee80216bit", "Ieee80216bit"},
	{"Uid", "Uid", "UID"},
	{"UUId", "Uuid", "UUID"},
	{"Uid_121_abd", "Uid121Abd", "UID121Abd"},
	{"a_UUId_b", "AUuidB", "AUUIDB"},
	{"AAA__Uid", "AaaUid", "AaaUID"},
	{"AA_DDD_UUId_12", "AaDddUuid12", "AaDddUUID12"},
}

func TestAll(t *testing.T) {
	TestToGoNames(t)
}

func TestToGoNames(t *testing.T) {
	for _, words := range toGoNamesTestDatas {
		a, b := toGoNames(words[0])
		if a != words[1] {
			t.Errorf("普通方式转换错误,\"%s\"->\"%s\"", words[0], a)
		}
		if b != words[2] {
			t.Errorf("特珠字符大写方式转换错误,\"%s\"->\"%s\"", words[0], b)
		}
	}
}
