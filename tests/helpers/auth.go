// auth.go
//
// A scalable, high performance drop-in replacement for the jam-build nodejs data service
// Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC
//
// This file is part of jam-build-propsdb.
// jam-build-propsdb is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later version.
// jam-build-propsdb is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
// You should have received a copy of the GNU Affero General Public License along with jam-build-propsdb.
// If not, see <https://www.gnu.org/licenses/>.
// Additional terms under GNU AGPL version 3 section 7:
// a) The reasonable legal notice of original copyright and author attribution must be preserved
//    by including the string: "Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC"
//    in this material, copies, or source code of derived works.

package helpers

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/localnerve/authorizer-go"
)

func randInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

// GeneratePassword generates a 10 character password with a capital and special char
func GeneratePassword() string {
	const (
		lower   = "abcdefghijklmnopqrstuvwxyz"
		upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		special = "!@#$%^&*"
		numbers = "0123456789"
		all     = lower + upper + special + numbers
	)

	password := make([]byte, 10)
	password[0] = upper[randInt(len(upper))]
	password[1] = special[randInt(len(special))]
	password[2] = numbers[randInt(len(numbers))]

	for i := 3; i < 10; i++ {
		password[i] = all[randInt(len(all))]
	}

	for i := range password {
		j := randInt(len(password))
		password[i], password[j] = password[j], password[i]
	}

	return string(password)
}

// AcquireAccount performs signup and login to get an access token
func AcquireAccount(t *testing.T, authzURL, email, password string, roles []string) string {
	client, err := authorizer.NewAuthorizerClient("test_client", authzURL, "", nil)
	if err != nil {
		t.Fatalf("Failed to create authorizer client: %v", err)
	}

	// Convert roles to []*string if needed
	rolesPtrs := make([]*string, len(roles))
	for i := range roles {
		rolesPtrs[i] = &roles[i]
	}

	// Signup
	signupReq := &authorizer.SignUpInput{
		Email:           &email,
		Password:        password,
		ConfirmPassword: password,
		Roles:           rolesPtrs,
	}

	_, err = client.SignUp(signupReq)
	if err != nil {
		// If user already exists, we might ignore error and try login
		t.Logf("Signup failed (might already exist): %v", err)
	}

	// Login
	loginReq := &authorizer.LoginInput{
		Email:    &email,
		Password: password,
	}

	res, err := client.Login(loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if res.AccessToken == nil {
		t.Fatal("Access token is nil")
	}

	return *res.AccessToken
}
