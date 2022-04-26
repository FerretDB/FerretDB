package integration

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
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
	}

	testCases := map[string]map[string]string{
		"linuxMintOsRelease": {
			"NAME":               "Linux Mint",
			"VERSION":            "20.3 (Una)",
			"ID":                 "linuxmint",
			"ID_LIKE":            "ubuntu",
			"PRETTY_NAME":        "Linux Mint 20.3",
			"VERSION_ID":         "20.3",
			"HOME_URL":           "https://www.linuxmint.com/",
			"SUPPORT_URL":        "https://forums.linuxmint.com/",
			"BUG_REPORT_URL":     "http://linuxmint-troubleshooting-guide.readthedocs.io/en/latest/",
			"PRIVACY_POLICY_URL": "https://www.linuxmint.com/",
			"VERSION_CODENAME":   "una",
			"UBUNTU_CODENAME":    "focal",
		},
		"ubuntuOsRelease": {
			"NAME":               "Ubuntu",
			"VERSION":            "18.04 LTS (Bionic Beaver)",
			"ID":                 "ubuntu",
			"ID_LIKE":            "debian",
			"PRETTY_NAME":        "Ubuntu 18.04 LTS",
			"VERSION_ID":         "18.04",
			"HOME_URL":           "https://www.ubuntu.com/",
			"SUPPORT_URL":        "https://help.ubuntu.com/",
			"BUG_REPORT_URL":     "https://bugs.launchpad.net/ubuntu/",
			"PRIVACY_POLICY_URL": "https://www.ubuntu.com/legal/terms-and-policies/privacy-policy",
			"VERSION_CODENAME":   "bionic",
			"UBUNTU_CODENAME":    "bionic",
		},
	}

	for key, testCase := range testCases {
		osReleaseReader := bytes.NewReader([]byte(osReleaseFilesData[key]))

		osRelease, err := common.ParseOSRelease(osReleaseReader)
		require.NoError(t, err)

		for testCaseKey, testCaseValue := range testCase {
			if testCaseValue != osRelease[testCaseKey] {
				t.Fatalf("values are not the same, actual: %s, expected: %s", testCaseValue, osRelease[testCaseKey])
			}
		}
	}
}
