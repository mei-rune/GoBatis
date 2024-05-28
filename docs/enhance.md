我们支持 mybatis 的一些功能外，还作了一些增强

## ${} 的改进

我们不支持 ${xxx}, 但是我提供了一个更安全的 \<print fmt="%s" value="b" inStr="true" /> 来替换它

     当  inStr="true" 时我会检查 value 的值中是不是有 引号之类的字符，防止 sql 注入
     
     当  inStr="false" 时我会检查 value 的值中是不是有 and 或  or 之类的逻辑表达式，防止 sql 注入


## 分页支持

   因为不同数据库的 offsetlimit 语法并不一致， 我们增加了 <pagination />, 它有两种格式

   1. \<pagination page="page" size="size" />
   2. \<pagination offset="offset" limit="limit" />

  
## sort 支持

   我们经常从web前端接收字段排序的要求，他们常常喜欢传 "+createdtime" 或 "-createdtime" 来表示按创建时间增序排行或按创建时间减序排行，所以我添加了 <sort_by /> 来实现这个功能

   \<sort_by sort="sort" />

   如果参数 sort 的值为 "+createdtime" 我们就将它转换为 order by createdtime asc
   如果参数 sort 的值为 "-createdtime" 我们就将它转换为 order by createdtime desc


## like 支持

   前台常常有模糊查询的需求， 但是他们常常忘了在值的前后加 % 号, 所以我添加了 <like />

   \<like value="keyword" isPrefix="true" isSuffix="true" />


   有这个表达式后， 在后台会对 keyword 的值进行检查
   如果 keyword 有 % 前缀或后缀则不作处理，否则 isPrefix=true 我会在 keyword 的前面添加 %, isSuffix=true 我会在 keyword 的后面添加 % ，如果 isPrefix 和 isSuffix 都为 false 时那么我们会在 keyword 的前后都加上 %。
 
## else 支持

   mybatis 的 \<if> 是不支持 else 的，它是希望你用 choose， 但我觉得不方便，我增加了  else 的支持， 如下

   \<if test="xxx"> ok \<else /> nok \</if>
