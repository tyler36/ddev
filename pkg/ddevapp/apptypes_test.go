package ddevapp_test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ddev/ddev/pkg/ddevapp"
	"github.com/ddev/ddev/pkg/fileutil"
	"github.com/ddev/ddev/pkg/nodeps"
	"github.com/ddev/ddev/pkg/testcommon"
	"github.com/ddev/ddev/pkg/util"
	asrt "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectAppType does a simple test of various filesystem setups to make
// sure the expected apptype is returned.
func TestDetectAppType(t *testing.T) {
	origDir, _ := os.Getwd()
	appTypes := ddevapp.GetValidAppTypes()
	var notSimplePHPAppTypes = []string{}
	for _, t := range appTypes {
		// we don't detect "generic", "php", "drupal" app types
		// They can't be "detected" anyway
		if t != nodeps.AppTypePHP && t != nodeps.AppTypeDrupal && t != nodeps.AppTypeGeneric {
			notSimplePHPAppTypes = append(notSimplePHPAppTypes, t)
		}
	}
	tmpDir := testcommon.CreateTmpDir(t.Name())

	t.Cleanup(func() {
		_ = os.Chdir(origDir)
		_ = os.RemoveAll(tmpDir)
	})

	err := fileutil.CopyDir(filepath.Join(origDir, "testdata", t.Name()), filepath.Join(tmpDir, "sampleapptypes"))
	require.NoError(t, err)
	for _, appType := range notSimplePHPAppTypes {
		app, err := ddevapp.NewApp(filepath.Join(tmpDir, "sampleapptypes", appType), true)
		require.NoError(t, err)
		app.Docroot = ddevapp.DiscoverDefaultDocroot(app)
		t.Cleanup(func() {
			_ = app.Stop(true, false)
		})

		foundType := app.DetectAppType()
		require.EqualValues(t, appType, foundType)

		// `generic` type should not be overridden
		app.Type = nodeps.AppTypeGeneric
		foundType = app.DetectAppType()
		require.EqualValues(t, nodeps.AppTypeGeneric, foundType)
	}
}

// TestConfigOverrideAction tests that the ConfigOverride action is properly applied, but only if the
// config is not included in the config.yaml.
func TestConfigOverrideAction(t *testing.T) {
	assert := asrt.New(t)
	origDir, _ := os.Getwd()

	appTypes := map[string]string{
		nodeps.AppTypeBackdrop:     nodeps.PHPDefault,
		nodeps.AppTypeCakePHP:      nodeps.PHPDefault,
		nodeps.AppTypeCraftCms:     nodeps.PHPDefault,
		nodeps.AppTypeDrupal6:      nodeps.PHP56,
		nodeps.AppTypeDrupal7:      nodeps.PHP82,
		nodeps.AppTypeDrupal11:     nodeps.PHPDefault,
		nodeps.AppTypeLaravel:      nodeps.PHPDefault,
		nodeps.AppTypeMagento:      nodeps.PHPDefault,
		nodeps.AppTypeMagento2:     nodeps.PHPDefault,
		nodeps.AppTypeSilverstripe: nodeps.PHPDefault,
		nodeps.AppTypeSymfony:      nodeps.PHPDefault,
		nodeps.AppTypeWordPress:    nodeps.PHPDefault,
	}

	for appType, expectedPHPVersion := range appTypes {
		testDir := testcommon.CreateTmpDir(t.Name())

		app, err := ddevapp.NewApp(testDir, true)
		assert.NoError(err)

		t.Cleanup(func() {
			err = os.Chdir(origDir)
			assert.NoError(err)
			err = app.Stop(true, false)
			assert.NoError(err)
			_ = os.RemoveAll(testDir)
		})

		// Prompt for apptype as a way to get it into the config.
		input := fmt.Sprintf("%s\n", appType)
		scanner := bufio.NewScanner(strings.NewReader(input))
		util.SetInputScanner(scanner)
		err = app.AppTypePrompt()
		assert.NoError(err)
		fmt.Println("")

		// With no config file written, the ConfigFileOverrideAction should result in an override
		err = app.ConfigFileOverrideAction(true)
		assert.NoError(err)

		// With a basic new app, the expectedPHPVersion should be the default
		assert.EqualValues(expectedPHPVersion, app.PHPVersion, "expected PHP version %s but got %s for apptype=%s", expectedPHPVersion, app.PHPVersion, appType)

		newVersion := "19.0-" + appType
		app.PHPVersion = newVersion
		err = app.WriteConfig()
		assert.NoError(err)
		err = app.ConfigFileOverrideAction(false)
		assert.NoError(err)
		// But with a config that has been written with a specified version, the version should be untouched by
		// app.ConfigFileOverrideAction()
		assert.EqualValues(app.PHPVersion, newVersion)
	}
}

// TestConfigOverrideActionOnExistingConfig tests that the ConfigOverride action is properly applied, even if the config
// existed when the override flag is enabled.
func TestConfigOverrideActionOnExistingConfig(t *testing.T) {
	assert := asrt.New(t)
	origDir, _ := os.Getwd()

	// This will only work for those project types defining configOverrideAction and altering php version
	appTypes := map[string]string{
		nodeps.AppTypeDrupal6: nodeps.PHP56,
		nodeps.AppTypeDrupal7: nodeps.PHP82,
		// For AppTypeDrupal we can't guess a version without a working installation.
	}

	for appType, expectedPHPVersion := range appTypes {
		testDir := testcommon.CreateTmpDir(t.Name())

		app, err := ddevapp.NewApp(testDir, true)
		assert.NoError(err)

		t.Cleanup(func() {
			err = os.Chdir(origDir)
			assert.NoError(err)
			err = app.Stop(true, false)
			assert.NoError(err)
			_ = os.RemoveAll(testDir)
		})

		// Prompt for apptype as a way to get it into the config.
		input := fmt.Sprintf("%s\n", appType)
		scanner := bufio.NewScanner(strings.NewReader(input))
		util.SetInputScanner(scanner)
		err = app.AppTypePrompt()
		assert.NoError(err)
		fmt.Println("")
		// We write the config first time.
		newVersion := "19.0-" + appType
		app.PHPVersion = newVersion
		err = app.WriteConfig()

		assert.EqualValues(newVersion, app.PHPVersion, "expected PHP version %s but got %s for apptype=%s", newVersion, app.PHPVersion, appType)

		err = app.ConfigFileOverrideAction(true)
		assert.NoError(err)
		// We can override existing config.
		assert.EqualValues(expectedPHPVersion, app.PHPVersion, "expected PHP version %s but got %s for apptype=%s", expectedPHPVersion, app.PHPVersion, appType)
	}
}
