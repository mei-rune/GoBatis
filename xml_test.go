package gobatis_test

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
)

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

	cfg := &gobatis.Config{DriverName: "postgres",
		DataSource: "aa",
		XMLPaths: []string{"tests",
			"../tests",
			"../../tests"},
		MaxIdleConns: 2,
		MaxOpenConns: 2,
		ShowSQL:      true,
		Logger:       log.New(os.Stdout, "[gobatis] ", log.Flags()),
	}

	var query *Query = nil
	initCtx := &gobatis.InitContext{Config: cfg,
		Logger:     cfg.Logger,
		Dialect:    gobatis.DbTypePostgres,
		Mapper:     gobatis.CreateMapper("", nil, nil),
		Statements: make(map[string]*gobatis.MappedStatement)}

	for idx, test := range []xmlCase{
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
	} {

		stmt, err := gobatis.NewMapppedStatement(initCtx, "ddd", gobatis.StatementTypeSelect, gobatis.ResultStruct, test.sql)
		if err != nil {
			t.Log("[", idx, "] ", test.name, ":", test.sql)
			t.Error(err)
			continue
		}

		ctx, err := gobatis.NewContext(initCtx.Dialect, initCtx.Mapper, test.paramNames, test.paramValues)
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
	cfg := &gobatis.Config{DriverName: "postgres",
		DataSource: "aa",
		XMLPaths: []string{"tests",
			"../tests",
			"../../tests"},
		MaxIdleConns: 2,
		MaxOpenConns: 2,
		ShowSQL:      true,
		Logger:       log.New(os.Stdout, "[gobatis] ", log.Flags()),
	}

	initCtx := &gobatis.InitContext{Config: cfg,
		Logger:     cfg.Logger,
		Dialect:    gobatis.DbTypePostgres,
		Mapper:     gobatis.CreateMapper("", nil, nil),
		Statements: make(map[string]*gobatis.MappedStatement)}

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
	} {
		stmt, err := gobatis.NewMapppedStatement(initCtx, "ddd", gobatis.StatementTypeSelect, gobatis.ResultStruct, test.sql)
		if err != nil {
			if !strings.Contains(err.Error(), test.err) {
				t.Log("[", idx, "] ", test.name, ":", test.sql)
				t.Error("except", test.err)
				t.Error("got   ", err)
			}
			continue
		}

		ctx, err := gobatis.NewContext(initCtx.Dialect, initCtx.Mapper, test.paramNames, test.paramValues)
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
