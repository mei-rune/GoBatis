## 方法引用


有时一个接口的方法可以引用另一个接口的方法，是很有用的， 可以如下
````go
  // @reference RoleDao.ListByUserID
  Roles(id int64) ([]Role, error)
````