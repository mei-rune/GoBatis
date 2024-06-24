package core

import (
	"testing"
)

func TestReplaceAndOr(t *testing.T) {
	for _, test := range []struct {
		txt    string
		result string
	}{
		{
			txt:    "c or b",
			result: "c || b",
		},
		{
			txt:    "a or b",
			result: "a || b",
		},
		{
			txt:    "a OR b",
			result: "a || b",
		},
		{
			txt:    "a  or  b",
			result: "a  ||  b",
		},
		{
			txt:    "a\tor b",
			result: "a\t|| b",
		},
		{
			txt:    "a or\tb",
			result: "a ||\tb",
		},
		{
			txt:    "a\tor\tb",
			result: "a\t||\tb",
		},
		{
			txt:    "a o b",
			result: "a o b",
		},
		{
			txt:    "a>3 or b>0",
			result: "a>3 || b>0",
		},
		{
			txt:    "a \"or\" b",
			result: "a \"or\" b",
		},

		{
			txt:    "a \"or\" or b",
			result: "a \"or\" || b",
		},

		{
			txt:    "c and b",
			result: "c && b",
		},
		{
			txt:    "c AND b",
			result: "c && b",
		},
		{
			txt:    "c And b",
			result: "c && b",
		},
		{
			txt:    "c aNd b",
			result: "c && b",
		},
		{
			txt:    "a and b",
			result: "a && b",
		},
		{
			txt:    "a  and  b",
			result: "a  &&  b",
		},
		{
			txt:    "a\tand b",
			result: "a\t&& b",
		},
		{
			txt:    "a and\tb",
			result: "a &&\tb",
		},
		{
			txt:    "a\tand\tb",
			result: "a\t&&\tb",
		},

		{
			txt:    "a \"and\" b",
			result: "a \"and\" b",
		},

		{
			txt:    "a \"and\" and b",
			result: "a \"and\" && b",
		},

		{
			txt:    "Area==8 and isNotEmpty(Name)",
			result: "Area==8 && isNotEmpty(Name)",
		},

		{
			txt:    "c gt b",
			result: "c > b",
		},
		{
			txt:    "a gt b",
			result: "a > b",
		},
		{
			txt:    "a  gt  b",
			result: "a  >  b",
		},
		{
			txt:    "a\tgt b",
			result: "a\t> b",
		},
		{
			txt:    "a gt\tb",
			result: "a >\tb",
		},
		{
			txt:    "a\tgt\tb",
			result: "a\t>\tb",
		},

		{
			txt:    "a \"gt\" b",
			result: "a \"gt\" b",
		},

		{
			txt:    "a \"gt\" gt b",
			result: "a \"gt\" > b",
		},

		{
			txt:    "c gte b",
			result: "c >= b",
		},
		{
			txt:    "a gte b",
			result: "a >= b",
		},
		{
			txt:    "a  gte  b",
			result: "a  >=  b",
		},
		{
			txt:    "a\tgte b",
			result: "a\t>= b",
		},
		{
			txt:    "a gte\tb",
			result: "a >=\tb",
		},
		{
			txt:    "a\tgte\tb",
			result: "a\t>=\tb",
		},

		{
			txt:    "a \"gte\" b",
			result: "a \"gte\" b",
		},

		{
			txt:    "a \"gte\" gte b",
			result: "a \"gte\" >= b",
		},

		{
			txt:    "c lt b",
			result: "c < b",
		},
		{
			txt:    "a lt b",
			result: "a < b",
		},
		{
			txt:    "a  lt  b",
			result: "a  <  b",
		},
		{
			txt:    "a\tlt b",
			result: "a\t< b",
		},
		{
			txt:    "a lt\tb",
			result: "a <\tb",
		},
		{
			txt:    "a\tlt\tb",
			result: "a\t<\tb",
		},

		{
			txt:    "a \"lt\" b",
			result: "a \"lt\" b",
		},

		{
			txt:    "a \"lt\" lt b",
			result: "a \"lt\" < b",
		},

		{
			txt:    "c lte b",
			result: "c <= b",
		},
		{
			txt:    "a lte b",
			result: "a <= b",
		},
		{
			txt:    "a  lte  b",
			result: "a  <=  b",
		},
		{
			txt:    "a\tlte b",
			result: "a\t<= b",
		},
		{
			txt:    "a lte\tb",
			result: "a <=\tb",
		},
		{
			txt:    "a\tlte\tb",
			result: "a\t<=\tb",
		},

		{
			txt:    "a \"lte\" b",
			result: "a \"lte\" b",
		},

		{
			txt:    "a \"lte\" lte b",
			result: "a \"lte\" <= b",
		},

		{
			txt:    "abc = \" a=b and c = 2 or a gt c or a lt b and a gte 9 or c lte 10 \"",
			result: "abc = \" a=b and c = 2 or a gt c or a lt b and a gte 9 or c lte 10 \"",
		},

		{
			txt:    "abc = ' a=b and c = 2 or a gt c or a lt b and a gte 9 or c lte 10 '",
			result: "abc = ' a=b and c = 2 or a gt c or a lt b and a gte 9 or c lte 10 '",
		},

		{
			txt:    "manufactor == 0",
			result: "manufactor == 0",
		},

		{
			txt:    "a",
			result: "a",
		},

		{
			txt:    "an",
			result: "an",
		},

		{
			txt:    "o",
			result: "o",
		},

		{
			txt:    "g",
			result: "g",
		},
		{
			txt:    "l",
			result: "l",
		},
		{
			txt:    "roleid > 0",
			result: "roleid > 0",
		},
		{
			txt:    "rorleid > 0",
			result: "rorleid > 0",
		},
		{
			txt:    "ranleid > 0",
			result: "ranleid > 0",
		},
		{
			txt:    "raleid > 0",
			result: "raleid > 0",
		},
		{
			txt:    "or > 0",
			result: "|| > 0",
		},
		{
			txt:    "orxxx > 0",
			result: "orxxx > 0",
		},
		{
			txt:    "and > 0",
			result: "&& > 0",
		},
		{
			txt:    "andsss > 0",
			result: "andsss > 0",
		},
		{
			txt:    "0 > an",
			result: "0 > an",
		},
		{
			txt:    "0 > o",
			result: "0 > o",
		},
		{
			txt:    "0 > l",
			result: "0 > l",
		},
		{
			txt:    "0 > g",
			result: "0 > g",
		},
		{
			txt:    "0 > gt",
			result: "0 > >",
		},
		{
			txt:    "0 > gte",
			result: "0 > >=",
		},
	} {
		s := replaceAndOr(test.txt)
		if s != test.result {
			t.Error("txt :", test.txt)
			t.Error("want:", test.result)
			t.Error(" got:", s)
		}
	}
}
