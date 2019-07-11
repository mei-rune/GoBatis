//go:generate gobatis upsert.go
package gentest

import "context"

type UserUpdater interface {
	UpsertByID(user *User) (int64, error)
	UpsertByIdByUsername(user *User) (int64, error)
	UpsertByUsername(user *User) (int64, error)

	// BD 为 BirthDay 的简写
	UpsertByBD(user *User) (int64, error)

	// User 为 UserID 的简写
	UpsertByUser(user *UserProfile) (int64, error)
	UpsertByKeyByUserID(user *UserProfile) (int64, error)
	// User 为 UserID 的简写
	UpsertByKeyByUser(user *UserProfile) (int64, error)
	// UI 为 UserID 的缩写
	UpsertByUID(user *UserProfile) (int64, error)

	// @record_type UserProfile
	UpsertXXX11(id int, user *UserProfile) (int64, error)

	// @record_type UserProfile
	UpsertXXX12(ctx context.Context, id int, user *UserProfile) (int64, error)

	// @record_type UserProfile
	UpsertXXX21(id int, username string, user *UserProfile) (int64, error)

	// @record_type UserProfile
	UpsertXXX22(ctx context.Context, id int, username string, user *UserProfile) (int64, error)
}
