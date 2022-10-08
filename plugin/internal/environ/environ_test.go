// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package environ

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSlice(t *testing.T) {
	v := map[string]string{
		"CI":    "true",
		"DRONE": "true",
	}
	a := Slice(v)
	b := []string{"CI=true", "DRONE=true"}
	if diff := cmp.Diff(a, b); diff != "" {
		t.Fail()
		t.Log(diff)
	}
}

func TestCombine(t *testing.T) {
	v1 := map[string]string{
		"CI":    "true",
		"DRONE": "true",
	}
	v2 := map[string]string{
		"CI":                    "false",
		"DRONE_SYSTEM_HOST":     "drone.company.com",
		"DRONE_SYSTEM_HOSTNAME": "drone.company.com",
		"DRONE_SYSTEM_PROTO":    "http",
		"DRONE_SYSTEM_VERSION":  "v1.0.0",
	}
	a := Combine(v1, v2)
	b := map[string]string{
		"CI":                    "false",
		"DRONE":                 "true",
		"DRONE_SYSTEM_HOST":     "drone.company.com",
		"DRONE_SYSTEM_HOSTNAME": "drone.company.com",
		"DRONE_SYSTEM_PROTO":    "http",
		"DRONE_SYSTEM_VERSION":  "v1.0.0",
	}
	if diff := cmp.Diff(a, b); diff != "" {
		t.Fail()
		t.Log(diff)
	}
}
