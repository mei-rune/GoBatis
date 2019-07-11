//go:generate gobatis upsert.go
package gentest

type UserUpdater interface {
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

	//	// @record_type UserProfile
	//	UpsertXXX11(id int, user *UserProfile) (int64, error)

	//	// @record_type UserProfile
	//	UpsertXXX12(ctx context.Context, id int, user *UserProfile) (int64, error)

	//	// @record_type UserProfile
	//	UpsertXXX21(id int, username string, user *UserProfile) (int64, error)

	//	// @record_type UserProfile
	//	UpsertXXX22(ctx context.Context, id int, username string, user *UserProfile) (int64, error)
}
