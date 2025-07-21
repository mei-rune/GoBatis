//go:generate gobatis upsert.go
package gentest

type Assoc1 struct {
	TableName struct{} `db:"assoc_table"`
	F1        int      `db:"f1,unique"`
	F2        int      `db:"f2,unique"`
}

type UserUpdater interface {
	// @record_type Assoc1
	UpsertAssoc1(f1, f2 int) error
	UpsertOnID(user *User) (int64, error)
	UpsertOnIdOnUsername(user *User) (int64, error)
	UpsertOnUsername(user *User) (int64, error)

	// BD 为 BirthDay 的简写
	UpsertOnBD(user *User) (int64, error)

	// User 为 UserID 的简写
	UpsertOnUser(user *UserProfile) (int64, error)
	UpsertOnKeyOnUserID(user *UserProfile) (int64, error)
	// User 为 UserID 的简写
	UpsertOnKeyOnUser(user *UserProfile) (int64, error)
	// UI 为 UserID 的缩写
	UpsertOnUID(user *UserProfile) (int64, error)

	// User 为 UserID 的简写
	Upsert2OnUser(user *UserProfile2) (int64, error)
	Upsert2OnKeyOnUserID(user *UserProfile2) (int64, error)
	// User 为 UserID 的简写
	Upsert2OnKeyOnUser(user *UserProfile2) (int64, error)
	// UI 为 UserID 的缩写
	Upsert2OnUID(user *UserProfile2) (int64, error)

	// User 为 UserID 的简写
	Upsert3OnUser(user *UserProfile3) (int64, error)
	Upsert3OnKeyOnUserID(user *UserProfile3) (int64, error)
	// User 为 UserID 的简写
	Upsert3OnKeyOnUser(user *UserProfile3) (int64, error)
	// UI 为 UserID 的缩写
	Upsert3OnUID(user *UserProfile3) (int64, error)

	//	// @record_type UserProfile
	//	UpsertXXX11(id int, user *UserProfile) (int64, error)

	//	// @record_type UserProfile
	//	UpsertXXX12(ctx context.Context, id int, user *UserProfile) (int64, error)

	//	// @record_type UserProfile
	//	UpsertXXX21(id int, username string, user *UserProfile) (int64, error)

	//	// @record_type UserProfile
	//	UpsertXXX22(ctx context.Context, id int, username string, user *UserProfile) (int64, error)
}
