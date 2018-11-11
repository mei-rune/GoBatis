# SQL 自动生成

和 MyBatis 不同，我觉得能像大部份的 orm 一样能生成 sql 的话，可以省很多工作.


生成规则如下
1. 生成 sql 不是很智能，只能生成一些针对一张表的简单操作，复杂的查询生成不了，请自已写 sql.
2. 接口中的每个方法都是对应 sql 的 insert, update, delete 和 select 四种中的一个，请见[SQL 配置](sql_config.md)，所以我们也只会生成这四种语句
3. 表名是根据配置中的 record_type 来获取的，或返回值的类型来获取的。

