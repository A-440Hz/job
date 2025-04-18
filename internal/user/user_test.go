package user

// test user CRUD
// user promotion/registration
// user cookies
import (
	"job/internal/db"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_validateRegisterBaseUser(t *testing.T) {
	db.SetEnvForTesting()
	db, err := db.InitGormDB()
	require.NoError(t, err)
	db.AutoMigrate(&User{})
	db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

	repo := NewRepository(db)
	svc := NewService(repo)

	tests := []struct {
		name       string
		username   string
		password   string
		email      string
		wantErrMsg []string
	}{
		{
			name:     "valid-fields",
			username: "u2",
			password: "p2",
			email:    "e2@mail.com",
		},
		{
			name:       "taken-username",
			username:   "u1",
			password:   "p2",
			email:      "e2@mail.com",
			wantErrMsg: []string{"username already registered"},
		},
		{
			name:       "taken-email",
			username:   "u2",
			password:   "p2",
			email:      "e1@mail.com",
			wantErrMsg: []string{"email already registered"},
		},
		{
			name:       "taken-email-password",
			username:   "u1",
			password:   "p2",
			email:      "e1@mail.com",
			wantErrMsg: []string{"username already registered", "email already registered"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// a constant user to test against
			u1, err := svc.CreateNewUser(nil)
			require.NoError(t, err)
			err = svc.RegisterBaseUser(u1.GetID(), "u1", "p1", "e1@mail.com")
			require.NoError(t, err)
			assert.NotNil(t, u1)

			// a new user to test against
			u2, err := svc.CreateNewUser(nil)
			assert.NoError(t, err)
			err = svc.RegisterBaseUser(u2.GetID(), tt.username, tt.password, tt.email)
			for _, msg := range tt.wantErrMsg {
				require.Error(t, err)
				assert.Contains(t, err.Error(), msg)
			}
			if len(tt.wantErrMsg) == 0 {
				assert.NotEqual(t, u1.GetID(), u2.GetID())
				assert.NotNil(t, u2)
			}
			db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
		})
	}
}

func Test_RegisterBaseUser(t *testing.T) {
	db.SetEnvForTesting()
	db, err := db.InitGormDB()
	require.NoError(t, err)
	db.AutoMigrate(&User{})
	db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

}

func Test_DeleteUser(t *testing.T) {
	db.SetEnvForTesting()
	db, err := db.InitGormDB()
	require.NoError(t, err)
	db.AutoMigrate(&User{})
	db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

	repo := NewRepository(db)
	svc := NewService(repo)

	u1, err := svc.CreateNewUser(nil)
	require.NoError(t, err)
	require.NotNil(t, u1)
	u2, err := svc.CreateNewUser(nil)
	require.NoError(t, err)
	require.NotNil(t, u2)
	u3, err := svc.CreateNewUser(nil)
	require.NoError(t, err)
	require.NotNil(t, u3)

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

	// delete 3 again
	err = svc.DeleteUser(id3)
	assert.Error(t, err)
	u, err = svc.LookupUser(id3)
	assert.ErrorContains(t, err, "record not found")
	assert.Nil(t, u)
	db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
}
