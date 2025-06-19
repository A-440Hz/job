package user

// test user CRUD
// user promotion/registration
// user cookies
import (
	"job/internal/collection"
	"job/internal/db"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func init() {
	db.ChdirGoTests()
}

// these are a reprecusion of using pointer attibutes in gorm
func int64Ptr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func Test_validateRegisterBaseUser(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, User{})
	svc := NewService(NewRepository(dBase), collection.NewService(collection.NewRepository(dBase)))

	tests := []struct {
		name       string
		Username   *string
		Password   *string
		Email      *string
		wantErrMsg []string
	}{
		{
			name:     "valid-fields",
			Username: strPtr("u2"),
			Password: strPtr("p2"),
			Email:    strPtr("e2@mail.com"),
		},
		{
			name:       "taken-username",
			Username:   strPtr("u1"),
			Password:   strPtr("p2"),
			Email:      strPtr("e2@mail.com"),
			wantErrMsg: []string{"username already taken"},
		},
		{
			name:       "taken-email",
			Username:   strPtr("u2"),
			Password:   strPtr("p2"),
			Email:      strPtr("e1@mail.com"),
			wantErrMsg: []string{"email already registered"},
		},
		{
			name:       "taken-email-password",
			Username:   strPtr("u1"),
			Password:   strPtr("p2"),
			Email:      strPtr("e1@mail.com"),
			wantErrMsg: []string{"username already taken", "email already registered"},
		},
		{
			name:       "missing-username",
			Password:   strPtr("p2"),
			Email:      strPtr("e2@mail.com"),
			wantErrMsg: []string{"missing username"},
		},
		{
			name:     "missing-email",
			Username: strPtr("u2"),
			Password: strPtr("p2"),
			// wantErrMsg: []string{"missing email"},
		},
		{
			name:       "missing-password",
			Username:   strPtr("u2"),
			Email:      strPtr("e2@mail.com"),
			wantErrMsg: []string{"missing password"},
		},
		{
			name:       "missing-all",
			wantErrMsg: []string{"missing username", "missing password"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// a constant user to test against
			defer db.CleanDB(*dBase, &User{})
			u1, err := svc.CreateNewUser(nil)
			require.NoError(t, err)
			u1, err = svc.RegisterBaseUser(u1.GetID(), &UserUpdateFields{
				Username:   strPtr("u1"),
				Password:   strPtr("p1"),
				Email:      strPtr("e1@mail.com"),
				Registered: boolPtr(true),
			})
			require.NoError(t, err)
			assert.NotNil(t, u1)

			// a new user to test against
			u2, err := svc.CreateNewUser(nil)
			assert.NoError(t, err)
			u2, err = svc.RegisterBaseUser(u2.GetID(), &UserUpdateFields{
				Username:   tt.Username,
				Password:   tt.Password,
				Email:      tt.Email,
				Registered: boolPtr(true),
			})
			for _, msg := range tt.wantErrMsg {
				assert.Contains(t, err.Error(), msg)
			}
			if len(tt.wantErrMsg) == 0 {
				assert.NotEqual(t, u1.GetID(), u2.GetID())
				assert.NotNil(t, u2)
			}
		})
	}
}

func Test_RegisterBaseUser(t *testing.T) {
	db.SetEnvForTesting()
	dbase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dbase, User{})
	repo := NewRepository(dbase)
	svc := NewService(repo, collection.NewService(collection.NewRepository(dbase)))

	u1, err := svc.CreateNewUser(nil)
	require.NoError(t, err)
	u1, err = svc.RegisterBaseUser(u1.GetID(), &UserUpdateFields{
		Username:   strPtr("u1"),
		Password:   strPtr("p1"),
		Email:      strPtr("e1@mail.com"),
		Registered: boolPtr(true),
	})
	require.NoError(t, err)
	assert.NotNil(t, u1)

	lookupU1, err := svc.LookupUser(u1.GetID())
	require.NoError(t, err)
	assert.NotNil(t, lookupU1)
	assert.Equal(t, u1.GetID(), lookupU1.GetID())
	assert.Equal(t, u1.Username, lookupU1.Username)
	assert.Equal(t, u1.Email, lookupU1.Email)
	assert.True(t, db.PasswordMatchesHash("p1", *lookupU1.Password))
	assert.True(t, lookupU1.Registered)
	assert.NotNil(t, lookupU1.Timezone)

	_, err = svc.RegisterBaseUser(u1.GetID(), &UserUpdateFields{
		Username:   strPtr("u1"),
		Password:   strPtr("p1"),
		Email:      strPtr("e1@mail.com"),
		Registered: boolPtr(true),
	})
	assert.ErrorContains(t, err, "user already registered")
	dbase.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
}

func Test_UpdateUserFields(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, User{})
	repo := NewRepository(dBase)
	svc := NewService(repo, collection.NewService(collection.NewRepository(dBase)))

	tests := []struct {
		name       string
		Registered *bool
		Username   *string
		Email      *string
		Timezone   *int64
		wantErrMsg []string
	}{
		{
			name:       "valid-fields",
			Registered: boolPtr(true),
			Username:   strPtr("u2"),
			Email:      strPtr("e2@mail.com"),
			Timezone:   int64Ptr(-1 * 3600),
		},
		{
			name:       "no-fields",
			wantErrMsg: []string{"no fields to update"},
		},
		{
			name:       "username-taken",
			Registered: boolPtr(true),
			Username:   strPtr("u1"),
			Email:      strPtr("e2@mail.com"),
			Timezone:   int64Ptr(-1 * 3600),
			wantErrMsg: []string{"username already taken"},
		},
		{
			name:       "email-taken-&-trim-space",
			Registered: boolPtr(true),
			Username:   strPtr("  u2  "),
			Email:      strPtr("    E1@Mail.cOm   "),
			Timezone:   int64Ptr(-1 * 3600),
			wantErrMsg: []string{"email already registered"},
		},
		{
			name:     "maxint-timezone",
			Timezone: int64Ptr(int64(^uint64(0) >> 1)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// a constant user to claim some fields
			defer dBase.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
			u1, err := svc.CreateNewUser(nil)
			require.NoError(t, err)
			u1, err = svc.RegisterBaseUser(u1.GetID(), &UserUpdateFields{
				Username:   strPtr("u1"),
				Password:   strPtr("p1"),
				Email:      strPtr("e1@mail.com"),
				Registered: boolPtr(true),
			})
			require.NoError(t, err)
			assert.NotNil(t, u1)

			u2, err := svc.CreateNewUser(nil)
			require.NoError(t, err)
			require.NotNil(t, u2)
			uf := &UserUpdateFields{
				Registered: tt.Registered,
				Username:   tt.Username,
				Email:      tt.Email,
				Timezone:   tt.Timezone,
			}
			updateU2, err := svc.UpdateUserFields(u2.GetID(), uf)
			if len(tt.wantErrMsg) > 0 {
				for _, msg := range tt.wantErrMsg {
					assert.Contains(t, err.Error(), msg)
				}
				assert.Empty(t, updateU2)
				return
			}
			assert.NoError(t, err)
			assert.NotEqual(t, u2, updateU2)
			if tt.Registered != nil {
				assert.Equal(t, updateU2.Registered, *tt.Registered)
			}
			assert.Equal(t, updateU2.Username, tt.Username)
			assert.Equal(t, updateU2.Email, tt.Email)
			assert.Equal(t, updateU2.Timezone.GetOffset(), *tt.Timezone)
		})
	}
}

func Test_DeleteUser(t *testing.T) {
	db.SetEnvForTesting()
	dBase, err := db.InitGormTestDB()
	require.NoError(t, err)
	db.CleanDB(*dBase, User{}, Session{}, &collection.UserInventory{})
	svc := NewService(NewRepository(dBase), collection.NewService(collection.NewRepository(dBase)))

	u1, err := svc.CreateNewUser(nil)
	require.NoError(t, err)
	require.NotNil(t, u1)
	u2, err := svc.CreateNewUser(nil)
	require.NoError(t, err)
	require.NotNil(t, u2)
	u3, err := svc.CreateNewUser(nil)
	require.NoError(t, err)
	require.NotNil(t, u3)
	ss, err := svc.CreateNewSession(u1.GetID())
	require.NoError(t, err)
	require.NotNil(t, ss)
	_, err = svc.CreateNewSession(u2.GetID())
	require.NoError(t, err)
	// leave u3 without a session

	// delete 1 2 3 once
	id3 := u3.GetID()
	u, err := svc.LookupUser(id3)
	assert.NoError(t, err)
	assert.NotNil(t, u)
	err = svc.DeleteUser(id3)
	assert.NoError(t, err)
	err = svc.DeleteUser(u2.GetID())
	assert.NoError(t, err)
	err = svc.DeleteUser(u1.GetID())
	assert.NoError(t, err)

	// lookup user inventory, session
	ui, err := svc.collection.LookupUserInventory(u2.GetID())
	assert.Nil(t, ui)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	sn, err := svc.repo.lookupSession(ss.ID)
	assert.Nil(t, sn)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// delete 3 again
	err = svc.DeleteUser(id3)
	assert.Error(t, err)
	u, err = svc.LookupUser(id3)
	assert.ErrorContains(t, err, "record not found")
	assert.Nil(t, u)

	// update fails on lookup
	u, err = svc.UpdateUserFields(id3, &UserUpdateFields{Username: strPtr("test")})
	assert.ErrorContains(t, err, "record not found")
	assert.Nil(t, u)
	db.CleanDB(*dBase, &User{}, &Session{}, &collection.UserInventory{})
}
