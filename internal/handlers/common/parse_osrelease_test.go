// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOSRelease(t *testing.T) {
	t.Parallel()

	osReleaseFilesData := map[string]string{
		"linuxMintOsRelease": `NAME=Linux Mint
VERSION=20.3 (Una)
ID=linuxmint
ID_LIKE=ubuntu
PRETTY_NAME=Linux Mint 20.3
VERSION_ID=20.3
HOME_URL=https://www.linuxmint.com/
SUPPORT_URL=https://forums.linuxmint.com/
BUG_REPORT_URL=http://linuxmint-troubleshooting-guide.readthedocs.io/en/latest/
PRIVACY_POLICY_URL=https://www.linuxmint.com/
VERSION_CODENAME=una
UBUNTU_CODENAME=focal
`,
		"ubuntuOsRelease": `NAME=Ubuntu
VERSION=18.04 LTS (Bionic Beaver)
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME=Ubuntu 18.04 LTS
VERSION_ID=18.04
HOME_URL=https://www.ubuntu.com/
SUPPORT_URL=https://help.ubuntu.com/
BUG_REPORT_URL=https://bugs.launchpad.net/ubuntu/
PRIVACY_POLICY_URL=https://www.ubuntu.com/legal/terms-and-policies/privacy-policy
VERSION_CODENAME=bionic
UBUNTU_CODENAME=bionic
`,
		"redHat": `NAME=Red Hat Enterprise Linux Server
VERSION=7.6 (Maipo)
ID=rhel
ID_LIKE=fedora
VARIANT=Server
VARIANT_ID=server
VERSION_ID=7.6
PRETTY_NAME=Red Hat Enterprise Linux
ANSI_COLOR=0;31
CPE_NAME=cpe:/o:redhat:enterprise_linux:7.6:GA:server
HOME_URL=https://www.redhat.com/
BUG_REPORT_URL=https://bugzilla.redhat.com/

REDHAT_BUGZILLA_PRODUCT=Red Hat Enterprise Linux 7
REDHAT_BUGZILLA_PRODUCT_VERSION=7.6
REDHAT_SUPPORT_PRODUCT=Red Hat Enterprise Linux
REDHAT_SUPPORT_PRODUCT_VERSION=7.6
`,
		"Arch Linux": `NAME="Arch Linux"
PRETTY_NAME="Arch Linux"
ID=arch
BUILD_ID=rolling
ANSI_COLOR="38;2;23;147;209"
HOME_URL="https://archlinux.org/"
DOCUMENTATION_URL="https://wiki.archlinux.org/"
SUPPORT_URL="https://bbs.archlinux.org/"
BUG_REPORT_URL="https://bugs.archlinux.org/"
PRIVACY_POLICY_URL="https://terms.archlinux.org/docs/privacy-policy/"
LOGO=archlinux-logo`,
	}

	testCases := map[string]map[string]string{
		"linuxMintOsRelease": {
			"NAME":    "Linux Mint",
			"VERSION": "20.3 (Una)",
		},
		"ubuntuOsRelease": {
			"NAME":    "Ubuntu",
			"VERSION": "18.04 LTS (Bionic Beaver)",
		},
		"redHat": {
			"NAME":    "Red Hat Enterprise Linux Server",
			"VERSION": "7.6 (Maipo)",
		},
		"Arch Linux": {
			"NAME":    "Arch Linux",
			"VERSION": "",
		},
	}

	for key, testCase := range testCases {
		osReleaseReader := bytes.NewReader([]byte(osReleaseFilesData[key]))

		osName, osVersion, err := parseOSRelease(osReleaseReader)
		require.NoError(t, err)

		assert.Equal(t, testCase["NAME"], osName)
		assert.Equal(t, testCase["VERSION"], osVersion)
	}
}
