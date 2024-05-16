package core_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/runner-mei/GoBatis/core"
	"github.com/runner-mei/GoBatis/dialects"
	"github.com/runner-mei/GoBatis/tests"
)

type mytesttype uint16

func TestXMLFiles(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {
		tmp := filepath.Join(getGoBatis(), "tmp/xmlgen")

		err := factory.ToXMLFiles(tmp)
		if err != nil {
			t.Error(err)
			return
		}
	})
}

type xmlCase struct {
	name        string
	sql         string
	paramNames  []string
	paramValues []interface{}

	exceptedSQL     string
	execeptedParams []interface{}
	isUnsortable    bool
}

type Query struct {
	StartAt, EndAt           time.Time
	OverdueStart, OverdueEnd time.Time
}

type XmlEmbedStruct1 struct {
	Field int
}

type XmlEmbedStruct2 struct {
	Field XmlEmbedStruct1
}

type XmlStruct struct {
	Field XmlEmbedStruct2
}

func TestXmlOk(t *testing.T) {
	cfg := &core.Config{DriverName: "postgres",
		DataSource: "aa",
		XMLPaths: []string{"tests",
			"../tests",
			"../../tests"},
		MaxIdleConns: 2,
		MaxOpenConns: 2,
		//ShowSQL:      true,
		Tracer: core.StdLogger{Logger: log.New(os.Stdout, "[gobatis] ", log.Flags())},
	}

	var query *Query = nil
	initCtx := &core.InitContext{Config: cfg,
		// Tracer:     cfg.Tracer,
		Dialect:    dialects.Postgres,
		Mapper:     core.CreateMapper("", nil, nil),
		Statements: make(map[string]*core.MappedStatement)}

	for idx, test := range []xmlCase{
		//		{
		//			name:            "escape",
		//			sql:             `aa < 0 and <if test="isnull(a)">bb</if>`,
		//			paramNames:      []string{"a"},
		//			paramValues:     []interface{}{query},
		//			exceptedSQL:     "aa < 0",
		//			execeptedParams: []interface{}{},
		//		},
		{
			name:            "if isnull",
			sql:             `aa <if test="isnull(a)">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{query},
			exceptedSQL:     "aa bb",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if isnotnull",
			sql:             `aa <if test="isnotnull(a)">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{query},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if isnotnull",
			sql:             `aa <if test="!isnull(a)">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{query},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if ok",
			sql:             `aa <if test="a==1">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{1},
			exceptedSQL:     "aa bb",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if ok",
			sql:             `aa <if test="a==1">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{2},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if ok",
			sql:             `aa <if test="a==1">#{a}</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{1},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{1},
		},
		{
			name:            "if ok",
			sql:             `aa #{b} <if test="a==1">#{a}</if>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{1, 2},
			exceptedSQL:     "aa $1 $2",
			execeptedParams: []interface{}{2, 1},
		},
		{
			name:            "if a == nil",
			sql:             `aa #{b} <if test="a != nil">#{a}</if>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{nil, 2},
			exceptedSQL:     "aa $1 ",
			execeptedParams: []interface{}{2},
		},
		{
			name:            "if ok",
			sql:             `aa <where><if test="a==1">#{a}</if></where>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{1},
			exceptedSQL:     "aa  WHERE $1",
			execeptedParams: []interface{}{1},
		},
		{
			name:            "if &&",
			sql:             `aa <where><if test="!query.StartAt.IsZero &amp;&amp; !query.EndAt.IsZero">#{a}</if></where>`,
			paramNames:      []string{"a", "query"},
			paramValues:     []interface{}{1, &Query{StartAt: time.Now(), EndAt: time.Now()}},
			exceptedSQL:     "aa  WHERE $1",
			execeptedParams: []interface{}{1},
		},
		{
			name:            "if &&",
			sql:             `aa <where><if test="!query.StartAt.IsZero &amp;&amp; !query.EndAt.IsZero">#{a}</if></where>`,
			paramNames:      []string{"a", "query"},
			paramValues:     []interface{}{1, &Query{}},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if else true",
			sql:             `aa <if test="trueValue">#{a}<else/></if>`,
			paramNames:      []string{"a", "trueValue"},
			paramValues:     []interface{}{1, true},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{1},
		},
		{
			name:            "if else true",
			sql:             `aa <if test="trueValue">#{a}<else/>#{b}</if>`,
			paramNames:      []string{"a", "trueValue", "b"},
			paramValues:     []interface{}{1, true, 2},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{1},
		},
		{
			name:            "if else true",
			sql:             `aa <if test="trueValue"><else/>#{b}</if>`,
			paramNames:      []string{"a", "trueValue", "b"},
			paramValues:     []interface{}{1, true, 2},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if else false",
			sql:             `aa <if test="trueValue">#{a}<else/>#{b}</if>`,
			paramNames:      []string{"a", "trueValue", "b"},
			paramValues:     []interface{}{1, false, 2},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{2},
		},
		{
			name:            "if else false",
			sql:             `aa <if test="trueValue"><else/>#{b}</if>`,
			paramNames:      []string{"a", "trueValue", "b"},
			paramValues:     []interface{}{1, false, 2},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{2},
		},
		{
			name:            "if else false",
			sql:             `aa <if test="trueValue">#{a}<else/></if>`,
			paramNames:      []string{"a", "trueValue", "b"},
			paramValues:     []interface{}{1, false, 2},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "where ok",
			sql:             `aa <where><if test="a==1">#{a}</if><if test="b==2">#{b}</if></where>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{1},
			exceptedSQL:     "aa  WHERE $1",
			execeptedParams: []interface{}{1},
		},
		{
			name:            "where ok",
			sql:             `aa <where><if test="a==1">#{a}</if><if test="b==2">#{b}</if></where>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{1, 2},
			exceptedSQL:     "aa  WHERE $1$2",
			execeptedParams: []interface{}{1, 2},
		},
		{
			name:            "where trim end",
			sql:             `aa <where><if test="a==1">#{a} and </if><if test="b==2">#{b} and </if></where>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{1, 2},
			exceptedSQL:     "aa  WHERE $1 and $2 ",
			execeptedParams: []interface{}{1, 2},
		},
		{
			name:            "where trim end",
			sql:             `aa <where><if test="a==1">#{a} and </if><if test="b==2">#{b} AND </if></where>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{1, 2},
			exceptedSQL:     "aa  WHERE $1 and $2 ",
			execeptedParams: []interface{}{1, 2},
		},
		{
			name:            "where trim or",
			sql:             `aa <where><if test="a==1">#{a} or </if><if test="b==2">#{b} or </if></where>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{1, 2},
			exceptedSQL:     "aa  WHERE $1 or $2 ",
			execeptedParams: []interface{}{1, 2},
		},
		{
			name:            "where trim or",
			sql:             `aa <where><if test="a==1">#{a} or </if><if test="b==2">#{b} OR </if></where>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{1, 2},
			exceptedSQL:     "aa  WHERE $1 or $2 ",
			execeptedParams: []interface{}{1, 2},
		},
		{
			name:            "where notok",
			sql:             `aa <where><if test="a==1">#{a}</if><if test="b==2">#{b}</if></where>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "where only and",
			sql:             `aa <where>and</where>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "where only and",
			sql:             `aa <where> and </where>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "where only or",
			sql:             `aa <where>or</where>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "where only or",
			sql:             `aa <where> or </where>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "where and",
			sql:             `aa <where>and abc</where>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa  WHERE  abc",
			execeptedParams: []interface{}{},
		},
		{
			name:            "where and",
			sql:             `aa <where> and abc </where>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa  WHERE  abc",
			execeptedParams: []interface{}{},
		},
		{
			name:            "where and",
			sql:             `aa <where> and abc and</where>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa  WHERE  abc ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "where only or",
			sql:             `aa <where>or abc</where>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa  WHERE  abc",
			execeptedParams: []interface{}{},
		},
		{
			name:            "where only or",
			sql:             `aa <where>or abc or</where>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa  WHERE  abc ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "where only or",
			sql:             `aa <where> or abc</where>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa  WHERE  abc",
			execeptedParams: []interface{}{},
		},
		{
			name:            "foreach ok",
			sql:             `aa <foreach collection="aa" index="index" item="item" open="(" separator="," close=")">#{item}</foreach>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{[]interface{}{"a", "b", "c", "d"}},
			exceptedSQL:     "aa ($1,$2,$3,$4)",
			execeptedParams: []interface{}{"a", "b", "c", "d"},
		},
		{
			name:            "foreach []string ok",
			sql:             `aa <foreach collection="aa" index="index" item="item" open="(" separator="," close=")">#{item}</foreach>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{[]string{"a", "b", "c", "d"}},
			exceptedSQL:     "aa ($1,$2,$3,$4)",
			execeptedParams: []interface{}{"a", "b", "c", "d"},
		},
		{
			name:            "foreach []int ok",
			sql:             `aa <foreach collection="aa" index="index" item="item" open="(" separator="," close=")">#{item}</foreach>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{[]int{22, 33, 44, 55}},
			exceptedSQL:     "aa ($1,$2,$3,$4)",
			execeptedParams: []interface{}{22, 33, 44, 55},
		},
		{
			name:            "foreach []int64 ok",
			sql:             `aa <foreach collection="aa" index="index" item="item" open="(" separator="," close=")">#{item}</foreach>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{[]int64{22, 33, 44, 55}},
			exceptedSQL:     "aa ($1,$2,$3,$4)",
			execeptedParams: []interface{}{int64(22), int64(33), int64(44), int64(55)},
		},
		{
			name:            "foreach []uint ok",
			sql:             `aa <foreach collection="aa" index="index" item="item" open="(" separator="," close=")">#{item}</foreach>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{[]uint{22, 33, 44, 55}},
			exceptedSQL:     "aa ($1,$2,$3,$4)",
			execeptedParams: []interface{}{uint(22), uint(33), uint(44), uint(55)},
		},
		{
			name:            "foreach collection not exists",
			sql:             `aa <foreach collection="abc" index="index" item="item" open="(" separator="," close=")">#{item}</foreach>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{[]interface{}{"a", "b", "c", "d"}},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "foreach array index",
			sql:             `aa <foreach collection="aa" index="index" item="item" open="(" separator="," close=")">#{index}</foreach>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{[]interface{}{"a", "b", "c", "d"}},
			exceptedSQL:     "aa ($1,$2,$3,$4)",
			execeptedParams: []interface{}{0, 1, 2, 3},
		},
		{
			name:            "foreach array index",
			sql:             `aa <foreach collection="aa" open="(" separator="," close=")">#{index}</foreach>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{[]interface{}{"a", "b", "c", "d"}},
			exceptedSQL:     "aa ($1,$2,$3,$4)",
			execeptedParams: []interface{}{0, 1, 2, 3},
		},
		{
			name:            "foreach map index",
			sql:             `aa <foreach collection="aa" index="index" item="item" open="(" separator="," close=")">#{index}</foreach>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{map[string]interface{}{"a": 1, "b": 2, "c": 3, "d": 4}},
			exceptedSQL:     "aa ($1,$2,$3,$4)",
			execeptedParams: []interface{}{"a", "b", "c", "d"},
			isUnsortable:    true,
		},
		{
			name:            "foreach map value",
			sql:             `aa <foreach collection="aa" index="index" item="item" open="(" separator="," close=")">#{item}</foreach>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{map[string]interface{}{"1": "a", "2": "b", "3": "c", "4": "d"}},
			exceptedSQL:     "aa ($1,$2,$3,$4)",
			execeptedParams: []interface{}{"a", "b", "c", "d"},
			isUnsortable:    true,
		},
		{
			name:            "foreach map value",
			sql:             `aa <foreach collection="aa" index="index" item="item" open="(" separator="," close=")">#{item}</foreach>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{map[interface{}]interface{}{"1": "a", "2": "b", "3": "c", "4": "d"}},
			exceptedSQL:     "aa ($1,$2,$3,$4)",
			execeptedParams: []interface{}{"a", "b", "c", "d"},
			isUnsortable:    true,
		},
		{
			name:            "chose ok",
			sql:             `aa <chose><when test="a==1">#{i1}</when></chose>`,
			paramNames:      []string{"a", "i1"},
			paramValues:     []interface{}{1, "a"},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{"a"},
		},
		{
			name:            "chose *",
			sql:             `aa <chose><when test="a==&quot;*&quot;">#{i1}</when></chose>`,
			paramNames:      []string{"a", "i1"},
			paramValues:     []interface{}{"*", "a"},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{"a"},
		},
		{
			name:            "chose *",
			sql:             `aa <chose><when test="a==&quot;*&quot;"></when><otherwise>#{i1}</otherwise></chose>`,
			paramNames:      []string{"a", "i1"},
			paramValues:     []interface{}{"*", "a"},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "chose noteq",
			sql:             `aa <chose><when test="a==1">#{i1}</when></chose>`,
			paramNames:      []string{"a", "i1"},
			paramValues:     []interface{}{2, "a"},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "chose otherwise",
			sql:             `aa <chose><when test="a==1">#{i1}</when><otherwise>#{i2}</otherwise></chose>`,
			paramNames:      []string{"a", "i1", "i2"},
			paramValues:     []interface{}{2, "a", "b"},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{"b"},
		},
		{
			name:            "chose missing",
			sql:             `aa <chose><when test="a==1">#{i1}</when></chose>`,
			paramNames:      []string{"i1"},
			paramValues:     []interface{}{"a"},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "set ok",
			sql:             `aa <set><if test="a==1">#{a}</if><if test="b==2">#{b}</if></set>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{1},
			exceptedSQL:     "aa  SET $1",
			execeptedParams: []interface{}{1},
		},
		{
			name:            "set ok",
			sql:             `aa <set><if test="a==1">#{a}</if><if test="b==2">#{b}</if></set>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{1, 2},
			exceptedSQL:     "aa  SET $1$2",
			execeptedParams: []interface{}{1, 2},
		},
		{
			name:            "set notok",
			sql:             `aa <set><if test="a==1">#{a}</if><if test="b==2">#{b}</if></set>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name: "where notok",
			sql: `aa <where><chose>
							<when test="len(a)==0">0</when>
							<when test="len(a)==1">1</when>
							<otherwise>more</otherwise>
					</chose></where>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]int{}},
			exceptedSQL:     "aa  WHERE 0",
			execeptedParams: []interface{}{},
		},
		{
			name: "where notok",
			sql: `aa <where><chose>
							<when test="len(a)==0">0</when>
							<when test="len(a)==1">1</when>
							<otherwise>more</otherwise>
					</chose></where>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]int{1}},
			exceptedSQL:     "aa  WHERE 1",
			execeptedParams: []interface{}{},
		},
		{
			name: "where notok",
			sql: `aa <where><chose>
							<when test="len(a)==0">0</when>
							<when test="len(a)==1">1</when>
							<otherwise>more</otherwise>
					</chose></where>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]int{1, 3}},
			exceptedSQL:     "aa  WHERE more",
			execeptedParams: []interface{}{},
		},
		{
			name: "where notok",
			sql: `aa #{b} <where><chose>
							<when test="len(a)==0">0</when>
							<when test="len(a)==1">1</when>
							<otherwise>
							<foreach collection="a" open="in (" close=")" separator=",">#{item}</foreach>
							</otherwise>
					</chose></where>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{[]int{1, 3}, 2},
			exceptedSQL:     "aa $1  WHERE in ($2,$3)",
			execeptedParams: []interface{}{2, 1, 3},
		},
		{
			name: "where notok",
			sql: `aa #{b} <where><chose>
							<when test="len(a)==0"></when>
							<when test="len(a)==1">1</when>
							<otherwise>
							<foreach collection="a" open="in (" close=")" separator=",">#{item}</foreach>
							</otherwise>
					</chose></where>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{[]int{}, 2},
			exceptedSQL:     "aa $1 ",
			execeptedParams: []interface{}{2},
		},
		{
			name:            "print_simple",
			sql:             `aa <print value="b"/>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{[]int{}, 2},
			exceptedSQL:     "aa 2",
			execeptedParams: []interface{}{},
		},
		{
			name:            "print_fmt",
			sql:             `aa <print fmt="%T" value="b"/> <print fmt="%d" value="b"/>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{[]int{}, 2},
			exceptedSQL:     "aa int 2",
			execeptedParams: []interface{}{},
		},
		{
			name:            "print_with_field.field.field",
			sql:             `aa <print value="a.Field.Field.Field"/>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{&XmlStruct{Field: XmlEmbedStruct2{Field: XmlEmbedStruct1{Field: 33}}}, 2},
			exceptedSQL:     "aa 33",
			execeptedParams: []interface{}{},
		},
		{
			name:            "print_simple_prefix_and_suffix_1",
			sql:             `aa <if test="b == 2" >  <print value="a"/> </if>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{33, 2},
			exceptedSQL:     "aa   33 ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "print_simple_prefix_and_suffix_2",
			sql:             `aa <if test="b == 2" >  <print value="a"/> <if test="b == 2"></if></if>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{33, 2},
			exceptedSQL:     "aa   33 ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "print_simple_prefix_and_suffix_2",
			sql:             `aa '<print value="a" inStr="true"/>'`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{"a b"},
			exceptedSQL:     "aa 'a b'",
			execeptedParams: []interface{}{},
		},
		{
			name:            "like_simple",
			sql:             `aa <like value="b"/>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{[]int{}, 2},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{"%2%"},
		},
		{
			name:            "like_with_field.field.field",
			sql:             `aa <like value="a.Field.Field.Field"/>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{&XmlStruct{Field: XmlEmbedStruct2{Field: XmlEmbedStruct1{Field: 33}}}, 2},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{"%33%"},
		},
		{
			name:            "like_simple_prefix_and_suffix_1",
			sql:             `aa <if test="b == 2" >  <like value="a"/> </if>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{33, 2},
			exceptedSQL:     "aa   $1 ",
			execeptedParams: []interface{}{"%33%"},
		},
		{
			name:            "like_simple_prefix_and_suffix_2",
			sql:             `aa <if test="b == 2" >  <like value="a"/> <if test="b == 2"></if></if>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{33, 2},
			exceptedSQL:     "aa   $1 ",
			execeptedParams: []interface{}{"%33%"},
		},
		{
			name:            "like_isprefix",
			sql:             `aa <like value="a" isPrefix="true"/>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{33, 2},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{"33%"},
		},
		{
			name:            "like_isprefix",
			sql:             `aa <like value="a" isSuffix="true"/>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{33, 2},
			exceptedSQL:     "aa $1",
			execeptedParams: []interface{}{"%33"},
		},
		{
			name:            "pagination 1",
			sql:             `aa <pagination offset="a" limit="b" />`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{33, 2},
			exceptedSQL:     "aa  OFFSET 33 LIMIT 2 ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "pagination 2",
			sql:             `aa <pagination />`,
			paramNames:      []string{"offset", "limit"},
			paramValues:     []interface{}{33, 2},
			exceptedSQL:     "aa  OFFSET 33 LIMIT 2 ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "pagination 3",
			sql:             `aa <pagination />`,
			paramNames:      []string{"offset", "limit"},
			paramValues:     []interface{}{0, 0},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "pagination 4",
			sql:             `aa <pagination />`,
			paramNames:      []string{"offset", "limit"},
			paramValues:     []interface{}{1, 0},
			exceptedSQL:     "aa  OFFSET 1 ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "pagination 5",
			sql:             `aa <pagination />`,
			paramNames:      []string{"offset", "limit"},
			paramValues:     []interface{}{0, 1},
			exceptedSQL:     "aa  LIMIT 1 ",
			execeptedParams: []interface{}{},
		},

		{
			name:            "pagination 1_1",
			sql:             `aa <pagination page="a" size="b" />`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{3, 2},
			exceptedSQL:     "aa  OFFSET 4 LIMIT 2 ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "pagination 1_2",
			sql:             `aa <pagination page="a" />`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{1, 0},
			exceptedSQL:     "aa  LIMIT 20 ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "pagination 1_3",
			sql:             `aa <pagination page="a" />`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{0, 0},
			exceptedSQL:     "aa  LIMIT 20 ",
			execeptedParams: []interface{}{},
		},

		{
			name:            "order by 1",
			sql:             `aa <order_by />`,
			paramNames:      []string{"sort"},
			paramValues:     []interface{}{"abc"},
			exceptedSQL:     "aa  ORDER BY abc",
			execeptedParams: []interface{}{},
		},
		{
			name:            "order by 2",
			sql:             `aa <order_by />`,
			paramNames:      []string{"sort"},
			paramValues:     []interface{}{"+abc"},
			exceptedSQL:     "aa  ORDER BY abc ASC",
			execeptedParams: []interface{}{},
		},
		{
			name:            "order by 3",
			sql:             `aa <order_by />`,
			paramNames:      []string{"sort"},
			paramValues:     []interface{}{"-abc"},
			exceptedSQL:     "aa  ORDER BY abc DESC",
			execeptedParams: []interface{}{},
		},
		{
			name:            "order by 4",
			sql:             `aa <order_by sort="aa"/>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{"-abc"},
			exceptedSQL:     "aa  ORDER BY abc DESC",
			execeptedParams: []interface{}{},
		},

		{
			name:            "order by 4",
			sql:             `aa <order_by sort="aa"/>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{"-abc +ddd"},
			exceptedSQL:     "aa  ORDER BY abc DESC, ddd ASC",
			execeptedParams: []interface{}{},
		},

		{
			name:            "order by 4",
			sql:             `aa <order_by sort="aa"/>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{"-abc -ddd"},
			exceptedSQL:     "aa  ORDER BY abc DESC, ddd DESC",
			execeptedParams: []interface{}{},
		},

		{
			name:            "order by 4",
			sql:             `aa <order_by sort="aa"/>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{"+abc -ddd"},
			exceptedSQL:     "aa  ORDER BY abc ASC, ddd DESC",
			execeptedParams: []interface{}{},
		},

		{
			name:            "order by 4",
			sql:             `aa <order_by sort="aa"/>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{"+abc +ddd"},
			exceptedSQL:     "aa  ORDER BY abc ASC, ddd ASC",
			execeptedParams: []interface{}{},
		},

		{
			name:            "trim prefix 1",
			sql:             `aa <trim prefixOverrides=",">, a</trim>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa  a",
			execeptedParams: []interface{}{},
		},
		{
			name:            "trim suffix 1",
			sql:             `aa <trim suffixOverrides=",">a ,</trim>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa a ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "trim prefix 2",
			sql:             `aa <trim prefixOverrides=","> a</trim>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa  a",
			execeptedParams: []interface{}{},
		},
		{
			name:            "trim suffix 2",
			sql:             `aa <trim suffixOverrides=",">a </trim>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa a ",
			execeptedParams: []interface{}{},
		},
		{
			name:            "trim prefix 3",
			sql:             `aa <trim prefixOverrides="," prefix="s">, a</trim>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa s a",
			execeptedParams: []interface{}{},
		},
		{
			name:            "trim suffix 3",
			sql:             `aa <trim suffixOverrides="," suffix="s">a ,</trim>`,
			paramNames:      []string{"aa"},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa a s",
			execeptedParams: []interface{}{},
		},
		{
			name:            "trim prefix 4",
			sql:             `aa #{a} <trim prefixOverrides="," prefix="#{b}">, a</trim>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{"1", "2"},
			exceptedSQL:     "aa $1 $2 a",
			execeptedParams: []interface{}{"1", "2"},
		},

		{
			name:            "trim prefix 5",
			sql:             `aa #{a} <trim prefixOverrides="," prefix="#{b}"> , a</trim>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{"1", "2"},
			exceptedSQL:     "aa $1 $2 a",
			execeptedParams: []interface{}{"1", "2"},
		},


		{
			name:            "trim prefix 6",
			sql:             `aa #{a} <trim prefixOverrides="and" prefix="#{b}"> and a</trim>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{"1", "2"},
			exceptedSQL:     "aa $1 $2 a",
			execeptedParams: []interface{}{"1", "2"},
		},

		{
			name:            "trim suffix 4",
			sql:             `aa #{a} <trim suffixOverrides="," suffix="#{b}">a ,</trim>`,
			paramNames:      []string{"a", "b"},
			paramValues:     []interface{}{"1", "2"},
			exceptedSQL:     "aa $1 a $2",
			execeptedParams: []interface{}{"1", "2"},
		},

		{
			name:            "trim prefix 5",
			sql:             `aa #{a} <trim prefixOverrides="," prefix="#{b}">, #{c}</trim>`,
			paramNames:      []string{"a", "b", "c"},
			paramValues:     []interface{}{"1", "2", "3"},
			exceptedSQL:     "aa $1 $2 $3",
			execeptedParams: []interface{}{"1", "2", "3"},
		},
		{
			name:            "trim suffix 5",
			sql:             `aa #{a} <trim suffixOverrides="," suffix="#{c}">#{b} ,</trim>`,
			paramNames:      []string{"a", "b", "c"},
			paramValues:     []interface{}{"1", "2", "3"},
			exceptedSQL:     "aa $1 $2 $3",
			execeptedParams: []interface{}{"1", "2", "3"},
		},
		{
			name:            "trim suffix 6",
			sql:             `aa, <trim suffixOverrides=","> id,name,age,address,area,card_no, </trim>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa,  id,name,age,address,area,card_no",
			execeptedParams: []interface{}{},
		},
		{
			name:            "trim suffix 7",
			sql:             `aa, <trim suffixOverrides="and"> id,name,age,address,area,card_no and </trim>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     "aa,  id,name,age,address,area,card_no ",
			execeptedParams: []interface{}{},
		},

		
		{
			name:            "qoute 1",
			sql:             `aa '"abc"'`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     `aa '"abc"'`,
			execeptedParams: []interface{}{},
		},
		{
			name:            "qoute 1",
			sql:             `aa <if test="true">'"abc"'</if>`,
			paramNames:      []string{},
			paramValues:     []interface{}{},
			exceptedSQL:     `aa '"abc"'`,
			execeptedParams: []interface{}{},
		},

		{
			name:            "value-range 1",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]int{1, 2}},
			exceptedSQL:     "aa  a BETWEEN $1 AND $2",
			execeptedParams: []interface{}{1, 2},
		},

		{
			name:            "value-range 2",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]int{1}},
			exceptedSQL:     "aa  a >= $1",
			execeptedParams: []interface{}{1},
		},

		{
			name:            "value-range 3",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]int{}},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},

		{
			name:            "value-range unsupport type 1",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]mytesttype{mytesttype(1), mytesttype(2)}},
			exceptedSQL:     "aa  a BETWEEN $1 AND $2",
			execeptedParams: []interface{}{mytesttype(1), mytesttype(2)},
		},

		{
			name:            "value-range unsupport type 2",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]mytesttype{mytesttype(1)}},
			exceptedSQL:     "aa  a >= $1",
			execeptedParams: []interface{}{mytesttype(1)},
		},

		{
			name:            "value-range unsupport type 3",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]mytesttype{}},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},

		{
			name:            "value-range time 1",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]time.Time{time.Unix(11, 0), time.Unix(12, 0)}},
			exceptedSQL:     "aa  a BETWEEN $1 AND $2",
			execeptedParams: []interface{}{time.Unix(11, 0), time.Unix(12, 0)},
		},

		{
			name:            "value-range time 2",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]time.Time{time.Unix(11, 0), time.Time{}}},
			exceptedSQL:     "aa  a >= $1",
			execeptedParams: []interface{}{time.Unix(11, 0)},
		},

		{
			name:            "value-range time 3",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]time.Time{time.Time{}, time.Unix(12, 0)}},
			exceptedSQL:     "aa  a <= $1",
			execeptedParams: []interface{}{time.Unix(12, 0)},
		},

		{
			name:            "value-range time 4",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]time.Time{time.Time{}, time.Time{}}},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},

		{
			name:            "value-range time 5",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]time.Time{time.Unix(11, 0)}},
			exceptedSQL:     "aa  a >= $1",
			execeptedParams: []interface{}{time.Unix(11, 0)},
		},

		{
			name:            "value-range time 6",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]time.Time{}},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},

		{
			name:            "value-range nullint64 1",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]sql.NullInt64{sql.NullInt64{Valid: true, Int64: 11}, sql.NullInt64{Valid: true, Int64: 12}}},
			exceptedSQL:     "aa  a BETWEEN $1 AND $2",
			execeptedParams: []interface{}{int64(11), int64(12)},
		},

		{
			name:            "value-range nullint64 2",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]sql.NullInt64{sql.NullInt64{Valid: true, Int64: 11}, sql.NullInt64{}}},
			exceptedSQL:     "aa  a >= $1",
			execeptedParams: []interface{}{int64(11)},
		},

		{
			name:            "value-range nullint64 3",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]sql.NullInt64{sql.NullInt64{}, sql.NullInt64{Valid: true, Int64: 12}}},
			exceptedSQL:     "aa  a <= $1",
			execeptedParams: []interface{}{int64(12)},
		},

		{
			name:            "value-range nullint64 4",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]sql.NullInt64{sql.NullInt64{}, sql.NullInt64{}}},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},

		{
			name:            "value-range nullint64 5",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]sql.NullInt64{sql.NullInt64{Valid: true, Int64: 11}}},
			exceptedSQL:     "aa  a >= $1",
			execeptedParams: []interface{}{int64(11)},
		},

		{
			name:            "value-range nullint64 6",
			sql:             `aa <value-range field="a" value="a" />`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{[]sql.NullInt64{}},
			exceptedSQL:     "aa ",
			execeptedParams: []interface{}{},
		},
	} {
		stmt, err := core.NewMapppedStatement(initCtx, "ddd", core.StatementTypeSelect, core.ResultStruct, test.sql)
		if err != nil {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error(err)
			continue
		}
		stmt.SQLStrings()

		ctx, err := core.NewContext(initCtx.Config.Constants, initCtx.Dialect, initCtx.Mapper, test.paramNames, test.paramValues)
		if err != nil {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error(err)
			continue
		}

		sqlParams, err := stmt.GenerateSQLs(ctx)
		if err != nil {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error(err)
			continue
		}
		if len(sqlParams) != 1 {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error("want sql rows is 1 got", len(sqlParams))
			continue
		}
		sqlStr := sqlParams[0].SQL
		params := sqlParams[0].Params

		if sqlStr != test.exceptedSQL {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error("except", fmt.Sprintf("%q", test.exceptedSQL))
			t.Error("got   ", fmt.Sprintf("%q", sqlStr))
			continue
		}

		if len(params) != 0 || len(test.execeptedParams) != 0 {

			var notOk = false
			if test.isUnsortable && len(params) == len(test.execeptedParams) {
				for idx := range params {
					found := false
					for _, a := range test.execeptedParams {
						if a == params[idx] {
							found = true
							break
						}
					}
					if !found {
						notOk = true
					}
				}
			} else if !reflect.DeepEqual(params, test.execeptedParams) {
				notOk = true
			}

			if notOk {
				t.Log("[", idx, "] ", test.name, ":", test.sql)
				t.Error("except", test.execeptedParams)
				t.Error("got   ", params)
				continue
			}
		}
	}

}

type xmlErrCase struct {
	name        string
	sql         string
	paramNames  []string
	paramValues []interface{}

	err string
}

func TestXmlFail(t *testing.T) {
	cfg := &core.Config{DriverName: "postgres",
		DataSource: "aa",
		XMLPaths: []string{"tests",
			"../tests",
			"../../tests"},
		MaxIdleConns: 2,
		MaxOpenConns: 2,
		//ShowSQL:      true,
		Tracer: core.StdLogger{Logger: log.New(os.Stdout, "[gobatis] ", log.Flags())},
	}

	initCtx := &core.InitContext{Config: cfg,
		// Logger:     cfg.Logger,
		Dialect:    dialects.Postgres,
		Mapper:     core.CreateMapper("", nil, nil),
		Statements: make(map[string]*core.MappedStatement)}

	for idx, test := range []xmlErrCase{
		{
			name:        "if ok",
			sql:         `aa <if >bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{1},
			err:         "is empty",
		},
		// {
		// 	name:        "if ok",
		// 	sql:         `aa <if test="a++"></if>`,
		// 	paramNames:  []string{"a"},
		// 	paramValues: []interface{}{1},
		// 	err:         "is empty",
		// },
		{
			name:        "if ok",
			sql:         `aa <if test="a+++">bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{1},
			err:         "+++",
		},
		{
			name:        "if ok",
			sql:         `aa <if test="a+1">bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{1},
			err:         "isnot bool",
		},
		{
			name:        "if ok",
			sql:         `aa <if test="a">bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{nil},
			err:         "is nil",
		},

		{
			name:        "if ok",
			sql:         `aa <chose><when >bb</when></chose>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{1},
			err:         "is empty",
		},

		{
			name:        "if ok",
			sql:         `aa <chose><otherwise>#{bb</otherwise></chose>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{1},
			err:         "#{bb",
		},

		{
			name:        "if ok",
			sql:         `aa <foreach>#{bb</foreach>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{1},
			err:         "#{bb",
		},

		{
			name:        "foreach execute",
			sql:         `aa <foreach collection="a" index="aa">#{item}</foreach>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{1},
			err:         "isnot slice",
		},

		{
			name:        "foreach bad argument",
			sql:         `aa <foreach collection="a" index="aa"></foreach>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{1},
			err:         "content",
		},

		{
			name:        "foreach bad argument 10",
			sql:         `aa <foreach>abc</foreach>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{1},
			err:         "collection",
		},

		{
			name:        "foreach bad argument",
			sql:         `aa <foreach collection="a" index="aa">#{</foreach>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{1},
			err:         "#{",
		},

		{
			name:        "like empty value",
			sql:         `aa <like value="a" />`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{""},
			err:         "empty",
		},

		{
			name:        "like xml error",
			sql:         `aa <like value="" />`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{""},
			err:         "has a 'value' notempty attribute",
		},

		{
			name:        "print xml error",
			sql:         `aa <print value="" />`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{""},
			err:         "has a 'value' notempty attribute",
		},
	} {
		stmt, err := core.NewMapppedStatement(initCtx, "ddd", core.StatementTypeSelect, core.ResultStruct, test.sql)
		if err != nil {
			if !strings.Contains(err.Error(), test.err) {
				t.Log("[", idx, "] ", test.name, ":", test.sql)
				t.Error("except", test.err)
				t.Error("got   ", err)
			}
			continue
		}
		stmt.SQLStrings()

		ctx, err := core.NewContext(initCtx.Config.Constants, initCtx.Dialect, initCtx.Mapper, test.paramNames, test.paramValues)
		if err != nil {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error(err)
			continue
		}

		_, err = stmt.GenerateSQLs(ctx)
		if err != nil {
			if !strings.Contains(err.Error(), test.err) {
				t.Log("[", idx, "] ", test.name, ":", test.sql)
				t.Error("except", test.err)
				t.Error("got   ", err)
			}
			continue
		}

		t.Log("[", idx, "] ", test.name, ":", test.sql)
		t.Error("except return a error")
		t.Error("got   ok")
	}

}

func TestXmlExpressionOk(t *testing.T) {
	cfg := &core.Config{DriverName: "postgres",
		DataSource: "aa",
		XMLPaths: []string{"tests",
			"../tests",
			"../../tests"},
		MaxIdleConns: 2,
		MaxOpenConns: 2,
		//ShowSQL:      true,
		Tracer: core.StdLogger{Logger: log.New(os.Stdout, "[gobatis] ", log.Flags())},
	}

	initCtx := &core.InitContext{Config: cfg,
		// Logger:     cfg.Logger,
		Dialect:    dialects.Postgres,
		Mapper:     core.CreateMapper("", nil, nil),
		Statements: make(map[string]*core.MappedStatement)}

	for idx, test := range []xmlCase{
		{
			name:            "if hasPrefix",
			sql:             `aa <if test="hasPrefix(a, &quot;a&quot;)">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{"abc"},
			exceptedSQL:     "aa bb",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if hasSuffix",
			sql:             `aa <if test="hasSuffix(a, &quot;a&quot;)">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{"bca"},
			exceptedSQL:     "aa bb",
			execeptedParams: []interface{}{},
		},

		{
			name:            "if hasSuffix",
			sql:             `aa <if test="trimPrefix(a, &quot;a&quot;) == &quot;bc&quot;">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{"abc"},
			exceptedSQL:     "aa bb",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if hasSuffix",
			sql:             `aa <if test="trimSuffix(a, &quot;a&quot;) == &quot;bc&quot;">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{"bca"},
			exceptedSQL:     "aa bb",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if hasSuffix",
			sql:             `aa <if test="trimSpace(a) == &quot;bc&quot;">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{" \t bc  \t \r \n"},
			exceptedSQL:     "aa bb",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if hasSuffix",
			sql:             `aa <if test="len(a) == 3">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{"abc"},
			exceptedSQL:     "aa bb",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if hasSuffix",
			sql:             `aa <if test="isEmpty(a)">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{""},
			exceptedSQL:     "aa bb",
			execeptedParams: []interface{}{},
		},
		{
			name:            "if hasSuffix",
			sql:             `aa <if test="isNotEmpty(a)">bb</if>`,
			paramNames:      []string{"a"},
			paramValues:     []interface{}{"a"},
			exceptedSQL:     "aa bb",
			execeptedParams: []interface{}{},
		},
	} {
		stmt, err := core.NewMapppedStatement(initCtx, "ddd", core.StatementTypeSelect, core.ResultStruct, test.sql)
		if err != nil {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error(err)
			continue
		}

		ctx, err := core.NewContext(initCtx.Config.Constants, initCtx.Dialect, initCtx.Mapper, test.paramNames, test.paramValues)
		if err != nil {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error(err)
			continue
		}

		sqlParams, err := stmt.GenerateSQLs(ctx)
		if err != nil {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error(err)
			continue
		}
		if len(sqlParams) != 1 {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error("want sql rows is 1 got", len(sqlParams))
			continue
		}
		sqlStr := sqlParams[0].SQL
		params := sqlParams[0].Params

		if sqlStr != test.exceptedSQL {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error("except", fmt.Sprintf("%q", test.exceptedSQL))
			t.Error("got   ", fmt.Sprintf("%q", sqlStr))
			continue
		}

		if len(params) != 0 || len(test.execeptedParams) != 0 {

			var notOk = false
			if test.isUnsortable && len(params) == len(test.execeptedParams) {
				for idx := range params {
					found := false
					for _, a := range test.execeptedParams {
						if a == params[idx] {
							found = true
							break
						}
					}
					if !found {
						notOk = true
					}
				}
			} else if !reflect.DeepEqual(params, test.execeptedParams) {
				notOk = true
			}

			if notOk {
				t.Log("[", idx, "] ", test.name, ":", test.sql)
				t.Error("except", test.execeptedParams)
				t.Error("got   ", params)
				continue
			}
		}
	}

}

func TestXmlExpressionFail(t *testing.T) {
	cfg := &core.Config{DriverName: "postgres",
		DataSource: "aa",
		XMLPaths: []string{"tests",
			"../tests",
			"../../tests"},
		MaxIdleConns: 2,
		MaxOpenConns: 2,
		//ShowSQL:      true,
		Tracer: core.StdLogger{Logger: log.New(os.Stdout, "[gobatis] ", log.Flags())},
	}

	initCtx := &core.InitContext{Config: cfg,
		// Logger:     cfg.Logger,
		Dialect:    dialects.Postgres,
		Mapper:     core.CreateMapper("", nil, nil),
		Statements: make(map[string]*core.MappedStatement)}

	for idx, test := range []xmlErrCase{
		{
			name:        "if hasPrefix",
			sql:         `aa <if test="hasPrefix(a)">bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{"abc"},
			err:         "invalid",
		},
		{
			name:        "if hasSuffix",
			sql:         `aa <if test="hasSuffix(a)">bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{"bca"},
			err:         "invalid",
		},
		{
			name:        "if trimPrefix",
			sql:         `aa <if test="trimPrefix(a) == &quot;bc&quot;">bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{"abc"},
			err:         "invalid",
		},
		{
			name:        "if trimSuffix",
			sql:         `aa <if test="trimSuffix(a) == &quot;bc&quot;">bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{"bca"},
			err:         "invalid",
		},
		{
			name:        "if trimSpace",
			sql:         `aa <if test="trimSpace(a, a) == &quot;bc&quot;">bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{" \t bc  \t \r \n"},
			err:         "invalid",
		},
		{
			name:        "if len",
			sql:         `aa <if test="len(a) == 3">bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{2},
			err:         "isnot",
		},
		{
			name:        "if isEmpty",
			sql:         `aa <if test="isEmpty(a, a)">bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{""},
			err:         "isnot",
		},
		{
			name:        "if hasSuffix",
			sql:         `aa <if test="isNotEmpty(a, a)">bb</if>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{"a"},
			err:         "isnot",
		},
		{
			name:        "else",
			sql:         `aa <else/>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{"a"},
			err:         "exist in the <if>",
		},
		{
			name:        "print_1",
			sql:         `aa <print value="a"/>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{"a b"},
			err:         core.ErrInvalidPrintValue.Error(),
		},
		{
			name:        "print_2",
			sql:         `aa <print value="a"/>`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{"a=b"},
			err:         core.ErrInvalidPrintValue.Error(),
		},
		{
			name:        "print_3",
			sql:         `aa '<print value="a" inStr="true"/>'`,
			paramNames:  []string{"a"},
			paramValues: []interface{}{"'"},
			err:         core.ErrInvalidPrintValue.Error(),
		},
	} {
		stmt, err := core.NewMapppedStatement(initCtx, "ddd", core.StatementTypeSelect, core.ResultStruct, test.sql)
		if err != nil {
			if !strings.Contains(err.Error(), test.err) {
				t.Log("[", idx, "] ", test.name, ":", test.sql)
				t.Error("except", test.err)
				t.Error("got   ", err)
			}
			continue
		}

		ctx, err := core.NewContext(initCtx.Config.Constants, initCtx.Dialect, initCtx.Mapper, test.paramNames, test.paramValues)
		if err != nil {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error(err)
			continue
		}

		_, err = stmt.GenerateSQLs(ctx)
		if err != nil {
			if !strings.Contains(err.Error(), test.err) {
				t.Log("[", idx, "] ", test.name, ":", test.sql)
				t.Error("except", test.err)
				t.Error("got   ", err)
			}
			continue
		}

		t.Log("[", idx, "] ", test.name, ":", test.sql)
		t.Error("except return a error")
		t.Error("got   ok")
	}
}
