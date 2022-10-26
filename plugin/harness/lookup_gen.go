// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

var index = map[string]plugin{
	"webhook": {
		name: "webhook",
		repo: "https://github.com/drone-plugins/drone-webhook.git",
		versions: map[string]string{
			"1.0.0": "b83c0042154f9c5d4bc3c42a847c3c287c12a505",
		},
	},
}
