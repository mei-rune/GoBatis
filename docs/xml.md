
 xml 中我们内建了下列函数

hasPrefix(s, prefix)  判断字符串 s 是否有 prefix 前缀
hasSuffix(s, suffix)  判断字符串 s 是否有 suffix 后缀
trimPrefix(s, prefix) 删除字符串 s 的 prefix 前缀（如果有的话）
trimSuffix(s, suffix) 删除字符串 s 的 suffix 后缀（如果有的话）
trimSpace(s)          删除字符串 s 前后的空格
len(slice)            获取 slice 的长度，它可以是 slice, array, map 和 str


isEmpty(s)            判断 s 是否为空，它可以是 slice, array, map 和 str
isNotEmpty(s)         判断 s 是否不为空，它可以是 slice, array, map 和 str

isEmptyString(s, isLike)            判断字符串 s 是否为空，isLike 为 true 时 ‘%’和‘%%’ 也作为空串
isZero(v)             判断 v 是否是不是一个零值
isNoZero(v)           判断 v 是否是不是一个非零值
isNull(v)             判断 v 是否是不是一个 nil 值

你也可以用 RegisterExprFunction 注册函数，注意你注册函数时必须在 gobatis.New() 调用之前，不然它将无效。
