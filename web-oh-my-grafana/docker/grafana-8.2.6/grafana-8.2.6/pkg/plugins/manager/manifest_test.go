package manager

import (
	"sort"
	"strings"
	"testing"

	"github.com/grafana/grafana/pkg/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadPluginManifest(t *testing.T) {
	txt := `-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA512

{
  "plugin": "grafana-googlesheets-datasource",
  "version": "1.0.0-dev",
  "files": {
    "LICENSE": "7df059597099bb7dcf25d2a9aedfaf4465f72d8d",
    "README.md": "08ec6d704b6115bef57710f6d7e866c050cb50ee",
    "gfx_sheets_darwin_amd64": "1b8ae92c6e80e502bb0bf2d0ae9d7223805993ab",
    "gfx_sheets_linux_amd64": "f39e0cc7344d3186b1052e6d356eecaf54d75b49",
    "gfx_sheets_windows_amd64.exe": "c8825dfec512c1c235244f7998ee95182f9968de",
    "module.js": "aaec6f51a995b7b843b843cd14041925274d960d",
    "module.js.LICENSE.txt": "7f822fe9341af8f82ad1b0c69aba957822a377cf",
    "module.js.map": "c5a524f5c4237f6ed6a016d43cd46938efeadb45",
    "plugin.json": "55556b845e91935cc48fae3aa67baf0f22694c3f"
  },
  "time": 1586817677115,
  "keyId": "7e4d0c6a708866e7"
}
-----BEGIN PGP SIGNATURE-----
Version: OpenPGP.js v4.10.1
Comment: https://openpgpjs.org

wqEEARMKAAYFAl6U6o0ACgkQfk0ManCIZuevWAIHSvcxOy1SvvL5gC+HpYyG
VbSsUvF2FsCoXUCTQflK6VdJfSPNzm8YdCdx7gNrBdly6HEs06ZaRp44F/ve
NR7DnB0CCQHO+4FlSPtXFTzNepoc+CytQyDAeOLMLmf2Tqhk2YShk+G/YlVX
74uuP5UXZxwK2YKJovdSknDIU7MhfuvvQIP/og==
=hBea
-----END PGP SIGNATURE-----`

	t.Run("valid manifest", func(t *testing.T) {
		manifest, err := readPluginManifest([]byte(txt))

		require.NoError(t, err)
		require.NotNil(t, manifest)
		assert.Equal(t, "grafana-googlesheets-datasource", manifest.Plugin)
		assert.Equal(t, "1.0.0-dev", manifest.Version)
		assert.Equal(t, int64(1586817677115), manifest.Time)
		assert.Equal(t, "7e4d0c6a708866e7", manifest.KeyID)
		expectedFiles := []string{"LICENSE", "README.md", "gfx_sheets_darwin_amd64", "gfx_sheets_linux_amd64",
			"gfx_sheets_windows_amd64.exe", "module.js", "module.js.LICENSE.txt", "module.js.map", "plugin.json",
		}
		assert.Equal(t, expectedFiles, fileList(manifest))
	})

	t.Run("invalid manifest", func(t *testing.T) {
		modified := strings.ReplaceAll(txt, "README.md", "xxxxxxxxxx")
		_, err := readPluginManifest([]byte(modified))
		require.Error(t, err)
	})
}

func TestReadPluginManifestV2(t *testing.T) {
	txt := `-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA512

{
  "manifestVersion": "2.0.0",
  "signatureType": "private",
  "signedByOrg": "willbrowne",
  "signedByOrgName": "Will Browne",
  "rootUrls": [
    "http://localhost:3000/"
  ],
  "plugin": "test",
  "version": "1.0.0",
  "time": 1605807018050,
  "keyId": "7e4d0c6a708866e7",
  "files": {
    "plugin.json": "2bb467c0bfd6c454551419efe475b8bf8573734e73c7bab52b14842adb62886f"
  }
}
-----BEGIN PGP SIGNATURE-----
Version: OpenPGP.js v4.10.1
Comment: https://openpgpjs.org

wqIEARMKAAYFAl+2q6oACgkQfk0ManCIZudmzwIJAXWz58cd/91rTXszKPnE
xbVEvERCbjKTtPBQBNQyqEvV+Ig3MuBSNOVy2SOGrMsdbS6lONgvgt4Cm+iS
wV+vYifkAgkBJtg/9DMB7/iX5O0h49CtSltcpfBFXlGqIeOwRac/yENzRzAA
khdr/tZ1PDgRxMqB/u+Vtbpl0xSxgblnrDOYMSI=
=rLIE
-----END PGP SIGNATURE-----`

	t.Run("valid manifest", func(t *testing.T) {
		manifest, err := readPluginManifest([]byte(txt))

		require.NoError(t, err)
		require.NotNil(t, manifest)
		assert.Equal(t, "test", manifest.Plugin)
		assert.Equal(t, "1.0.0", manifest.Version)
		assert.Equal(t, int64(1605807018050), manifest.Time)
		assert.Equal(t, "7e4d0c6a708866e7", manifest.KeyID)
		assert.Equal(t, "2.0.0", manifest.ManifestVersion)
		assert.Equal(t, plugins.PrivateType, manifest.SignatureType)
		assert.Equal(t, "willbrowne", manifest.SignedByOrg)
		assert.Equal(t, "Will Browne", manifest.SignedByOrgName)
		assert.Equal(t, []string{"http://localhost:3000/"}, manifest.RootURLs)
		assert.Equal(t, []string{"plugin.json"}, fileList(manifest))
	})
}

func fileList(manifest *pluginManifest) []string {
	var keys []string
	for k := range manifest.Files {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
