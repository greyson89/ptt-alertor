package user

import (
	"database/sql"
	"os"
	"reflect"
	"testing"

	"github.com/watain666/ptt-alertor/connections"
)

func TestMain(m *testing.M) {
	os.Setenv("SQLITE_PATH", ":memory:")
	connectDB = connections.DB

	os.Exit(m.Run())
}

func resetUsersTable(t *testing.T) *sql.DB {
	t.Helper()
	db := connectDB()
	if _, err := db.Exec("DELETE FROM users"); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestSQLite_List(t *testing.T) {
	db := resetUsersTable(t)
	db.Exec("INSERT INTO users (account, data) VALUES (?, ?)", "admin", `{"account":"admin"}`)

	tests := []struct {
		name         string
		r            SQLite
		wantAccounts []string
	}{
		{"admin", SQLite{}, []string{"admin"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAccounts := tt.r.List(); !reflect.DeepEqual(gotAccounts, tt.wantAccounts) {
				t.Errorf("SQLite.List() = %v, want %v", gotAccounts, tt.wantAccounts)
			}
		})
	}
}

func TestSQLite_Exist(t *testing.T) {
	db := resetUsersTable(t)
	db.Exec("INSERT INTO users (account, data) VALUES (?, ?)", "admin", `{"account":"admin"}`)

	type args struct {
		account string
	}
	tests := []struct {
		name string
		r    SQLite
		args args
		want bool
	}{
		{"admin", SQLite{}, args{"admin"}, true},
		{"missing", SQLite{}, args{"nobody"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Exist(tt.args.account); got != tt.want {
				t.Errorf("SQLite.Exist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLite_Save(t *testing.T) {
	resetUsersTable(t)

	type args struct {
		account string
		data    interface{}
	}
	tests := []struct {
		name    string
		r       SQLite
		args    args
		wantErr bool
	}{
		{"ok", SQLite{}, args{"admin", `{"account":"admin"}`}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Save(tt.args.account, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("SQLite.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSQLite_Update(t *testing.T) {
	db := resetUsersTable(t)
	db.Exec("INSERT INTO users (account, data) VALUES (?, ?)", "admin", `{"account":"admin"}`)

	type args struct {
		account string
		user    interface{}
	}
	tests := []struct {
		name    string
		r       SQLite
		args    args
		wantErr bool
	}{
		{"ok", SQLite{}, args{"admin", User{
			Enable: true,
			Profile: Profile{
				Account: "admin",
			}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Update(tt.args.account, tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("SQLite.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSQLite_Find(t *testing.T) {
	db := resetUsersTable(t)
	db.Exec("INSERT INTO users (account, data) VALUES (?, ?)", "admin", `{"Profile":{"account":"admin"}}`)

	type args struct {
		account string
		user    *User
	}
	tests := []struct {
		name string
		r    SQLite
		args args
	}{
		{"ok", SQLite{}, args{"admin", &User{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Find(tt.args.account, tt.args.user)
			if tt.args.user.Profile.Account != "admin" {
				t.Errorf("SQLite.Find() account = %v, want admin", tt.args.user.Profile.Account)
			}
		})
	}
}
