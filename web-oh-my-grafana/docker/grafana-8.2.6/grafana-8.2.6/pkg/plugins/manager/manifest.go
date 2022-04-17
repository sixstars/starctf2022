package manager

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util/errutil"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"
)

// Soon we can fetch keys from:
//  https://grafana.com/api/plugins/ci/keys
const publicKeyText = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: OpenPGP.js v4.10.1
Comment: https://openpgpjs.org

xpMEXpTXXxMFK4EEACMEIwQBiOUQhvGbDLvndE0fEXaR0908wXzPGFpf0P0Z
HJ06tsq+0higIYHp7WTNJVEZtcwoYLcPRGaa9OQqbUU63BEyZdgAkPTz3RFd
5+TkDWZizDcaVFhzbDd500yTwexrpIrdInwC/jrgs7Zy/15h8KA59XXUkdmT
YB6TR+OA9RKME+dCJozNGUdyYWZhbmEgPGVuZ0BncmFmYW5hLmNvbT7CvAQQ
EwoAIAUCXpTXXwYLCQcIAwIEFQgKAgQWAgEAAhkBAhsDAh4BAAoJEH5NDGpw
iGbnaWoCCQGQ3SQnCkRWrG6XrMkXOKfDTX2ow9fuoErN46BeKmLM4f1EkDZQ
Tpq3SE8+My8B5BIH3SOcBeKzi3S57JHGBdFA+wIJAYWMrJNIvw8GeXne+oUo
NzzACdvfqXAZEp/HFMQhCKfEoWGJE8d2YmwY2+3GufVRTI5lQnZOHLE8L/Vc
1S5MXESjzpcEXpTXXxIFK4EEACMEIwQBtHX/SD5Qm3v4V92qpaIZQgtTX0sT
cFPjYWAHqsQ1iENrYN/vg1wU3ADlYATvydOQYvkTyT/tbDvx2Fse8PL84MQA
YKKQ6AJ3gLVvmeouZdU03YoV4MYaT8KbnJUkZQZkqdz2riOlySNI9CG3oYmv
omjUAtzgAgnCcurfGLZkkMxlmY8DAQoJwqQEGBMKAAkFAl6U118CGwwACgkQ
fk0ManCIZuc0jAIJAVw2xdLr4ZQqPUhubrUyFcqlWoW8dQoQagwO8s8ubmby
KuLA9FWJkfuuRQr+O9gHkDVCez3aism7zmJBqIOi38aNAgjJ3bo6leSS2jR/
x5NqiKVi83tiXDPncDQYPymOnMhW0l7CVA7wj75HrFvvlRI/4MArlbsZ2tBn
N1c5v9v/4h6qeA==
=DNbR
-----END PGP PUBLIC KEY BLOCK-----
`

// pluginManifest holds details for the file manifest
type pluginManifest struct {
	Plugin  string            `json:"plugin"`
	Version string            `json:"version"`
	KeyID   string            `json:"keyId"`
	Time    int64             `json:"time"`
	Files   map[string]string `json:"files"`

	// V2 supported fields
	ManifestVersion string                      `json:"manifestVersion"`
	SignatureType   plugins.PluginSignatureType `json:"signatureType"`
	SignedByOrg     string                      `json:"signedByOrg"`
	SignedByOrgName string                      `json:"signedByOrgName"`
	RootURLs        []string                    `json:"rootUrls"`
}

func (m *pluginManifest) isV2() bool {
	return strings.HasPrefix(m.ManifestVersion, "2.")
}

// readPluginManifest attempts to read and verify the plugin manifest
// if any error occurs or the manifest is not valid, this will return an error
func readPluginManifest(body []byte) (*pluginManifest, error) {
	block, _ := clearsign.Decode(body)
	if block == nil {
		return nil, errors.New("unable to decode manifest")
	}

	// Convert to a well typed object
	manifest := &pluginManifest{}
	err := json.Unmarshal(block.Plaintext, &manifest)
	if err != nil {
		return nil, errutil.Wrap("Error parsing manifest JSON", err)
	}

	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(publicKeyText))
	if err != nil {
		return nil, errutil.Wrap("failed to parse public key", err)
	}

	if _, err := openpgp.CheckDetachedSignature(keyring,
		bytes.NewBuffer(block.Bytes),
		block.ArmoredSignature.Body); err != nil {
		return nil, errutil.Wrap("failed to check signature", err)
	}

	return manifest, nil
}

// getPluginSignatureState returns the signature state for a plugin.
func getPluginSignatureState(log log.Logger, plugin *plugins.PluginBase) (plugins.PluginSignatureState, error) {
	log.Debug("Getting signature state of plugin", "plugin", plugin.Id, "isBackend", plugin.Backend)
	manifestPath := filepath.Join(plugin.PluginDir, "MANIFEST.txt")

	// nolint:gosec
	// We can ignore the gosec G304 warning on this one because `manifestPath` is based
	// on plugin the folder structure on disk and not user input.
	byteValue, err := ioutil.ReadFile(manifestPath)
	if err != nil || len(byteValue) < 10 {
		log.Debug("Plugin is unsigned", "id", plugin.Id)
		return plugins.PluginSignatureState{
			Status: plugins.PluginSignatureUnsigned,
		}, nil
	}

	manifest, err := readPluginManifest(byteValue)
	if err != nil {
		log.Debug("Plugin signature invalid", "id", plugin.Id)
		return plugins.PluginSignatureState{
			Status: plugins.PluginSignatureInvalid,
		}, nil
	}

	// Make sure the versions all match
	if manifest.Plugin != plugin.Id || manifest.Version != plugin.Info.Version {
		return plugins.PluginSignatureState{
			Status: plugins.PluginSignatureModified,
		}, nil
	}

	// Validate that private is running within defined root URLs
	if manifest.SignatureType == plugins.PrivateType {
		appURL, err := url.Parse(setting.AppUrl)
		if err != nil {
			return plugins.PluginSignatureState{}, err
		}
		appSubURL, err := url.Parse(setting.AppSubUrl)
		if err != nil {
			return plugins.PluginSignatureState{}, err
		}
		appURLPath := path.Join(appSubURL.RequestURI(), appURL.RequestURI())

		foundMatch := false
		for _, u := range manifest.RootURLs {
			rootURL, err := url.Parse(u)
			if err != nil {
				log.Warn("Could not parse plugin root URL", "plugin", plugin.Id, "rootUrl", rootURL)
				return plugins.PluginSignatureState{}, err
			}

			if rootURL.Scheme == appURL.Scheme &&
				rootURL.Host == appURL.Host {
				foundMatch = path.Clean(rootURL.RequestURI()) == appURLPath

				if foundMatch {
					break
				}
			}
		}

		if !foundMatch {
			log.Warn("Could not find root URL that matches running application URL", "plugin", plugin.Id,
				"appUrl", appURL, "rootUrls", manifest.RootURLs)
			return plugins.PluginSignatureState{
				Status: plugins.PluginSignatureInvalid,
			}, nil
		}
	}

	manifestFiles := make(map[string]struct{}, len(manifest.Files))

	// Verify the manifest contents
	log.Debug("Verifying contents of plugin manifest", "plugin", plugin.Id)
	for fp, hash := range manifest.Files {
		err = verifyHash(plugin.Id, filepath.Join(plugin.PluginDir, fp), hash)
		if err != nil {
			return plugins.PluginSignatureState{
				Status: plugins.PluginSignatureModified,
			}, nil
		}

		manifestFiles[fp] = struct{}{}
	}

	if manifest.isV2() {
		pluginFiles, err := pluginFilesRequiringVerification(plugin)
		if err != nil {
			log.Warn("Could not collect plugin file information in directory", "pluginID", plugin.Id, "dir", plugin.PluginDir)
			return plugins.PluginSignatureState{
				Status: plugins.PluginSignatureInvalid,
			}, err
		}

		// Track files missing from the manifest
		var unsignedFiles []string
		for _, f := range pluginFiles {
			if _, exists := manifestFiles[f]; !exists {
				unsignedFiles = append(unsignedFiles, f)
			}
		}

		if len(unsignedFiles) > 0 {
			log.Warn("The following files were not included in the signature", "plugin", plugin.Id, "files", unsignedFiles)
			return plugins.PluginSignatureState{
				Status: plugins.PluginSignatureModified,
			}, nil
		}
	}

	// Everything OK
	log.Debug("Plugin signature valid", "id", plugin.Id)
	return plugins.PluginSignatureState{
		Status:     plugins.PluginSignatureValid,
		Type:       manifest.SignatureType,
		SigningOrg: manifest.SignedByOrgName,
		Files:      manifestFiles,
	}, nil
}

func verifyHash(pluginID string, path string, hash string) error {
	// nolint:gosec
	// We can ignore the gosec G304 warning on this one because `path` is based
	// on the path provided in a manifest file for a plugin and not user input.
	f, err := os.Open(path)
	if err != nil {
		log.Warn("Plugin file listed in the manifest was not found", "plugin", pluginID, "path", path)
		return fmt.Errorf("plugin file listed in the manifest was not found")
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Warn("Failed to close plugin file", "path", path, "err", err)
		}
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("could not calculate plugin file checksum")
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if sum != hash {
		log.Warn("Plugin file checksum does not match signature checksum", "plugin", pluginID, "path", path)
		return fmt.Errorf("plugin file checksum does not match signature checksum")
	}

	return nil
}

// gets plugin filenames that require verification for plugin signing
// returns filenames as a slice of posix style paths relative to plugin directory
func pluginFilesRequiringVerification(plugin *plugins.PluginBase) ([]string, error) {
	var files []string
	err := filepath.Walk(plugin.PluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			symlinkPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}

			symlink, err := os.Stat(symlinkPath)
			if err != nil {
				return err
			}

			// skip symlink directories
			if symlink.IsDir() {
				return nil
			}

			// verify that symlinked file is within plugin directory
			p, err := filepath.Rel(plugin.PluginDir, symlinkPath)
			if err != nil {
				return err
			}
			if strings.HasPrefix(p, ".."+string(filepath.Separator)) {
				return fmt.Errorf("file '%s' not inside of plugin directory", p)
			}
		}

		// skip directories and MANIFEST.txt
		if info.IsDir() || info.Name() == "MANIFEST.txt" {
			return nil
		}

		// verify that file is within plugin directory
		file, err := filepath.Rel(plugin.PluginDir, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(file, ".."+string(filepath.Separator)) {
			return fmt.Errorf("file '%s' not inside of plugin directory", file)
		}

		files = append(files, filepath.ToSlash(file))

		return nil
	})

	return files, err
}
