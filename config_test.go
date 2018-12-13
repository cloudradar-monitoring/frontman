package frontman

import (
	"io/ioutil"
	"os"
	"testing"

	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

func TestNewMinimumConfig(t *testing.T) {
	envURL := "http://foo.bar"
	envUser := "foo"
	envPass := "bar"

	// TODO: Not sure if this is really a good idea... could mess with other things
	os.Setenv("FRONTMAN_HUB_URL", envURL)
	os.Setenv("FRONTMAN_HUB_USER", envUser)
	os.Setenv("FRONTMAN_HUB_PASSWORD", envPass)

	mvc := NewMinimumConfig()

	assert.Equal(t, envURL, mvc.HubURL, "HubURL should be set from env")
	assert.Equal(t, envUser, mvc.HubUser, "HubUser should be set from env")
	assert.Equal(t, envPass, mvc.HubPassword, "HubPassword should be set from env")

	// Unset in the end for cleanup
	defer os.Clearenv()
}

func TestTryUpdateConfigFromFile(t *testing.T) {
	config := Config{
		PidFile:         "fooPID",
		Sleep:           0.1337,
		IgnoreSSLErrors: false,
	}

	const sampleConfig = `
pid = "/pid"
sleep = 1.0
ignore_ssl_errors = true
`

	tmpFile, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	defer os.Remove(tmpFile.Name())

	err = ioutil.WriteFile(tmpFile.Name(), []byte(sampleConfig), 0755)
	assert.Nil(t, err)

	err = TryUpdateConfigFromFile(&config, tmpFile.Name())
	assert.Nil(t, err)

	assert.Equal(t, "/pid", config.PidFile)
	assert.Equal(t, 1.0, config.Sleep)
	assert.Equal(t, true, config.IgnoreSSLErrors)
}

func TestGenerateDefaultConfigFile(t *testing.T) {
	mvc := &MinValuableConfig{
		LogLevel: "debug",
		IOMode:   "foo",
		HubUser:  "bar",
	}

	tmpFile, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	defer os.Remove(tmpFile.Name())

	err = GenerateDefaultConfigFile(mvc, tmpFile.Name())
	assert.Nil(t, err)

	loadedMVC := &MinValuableConfig{}
	err = toml.NewDecoder(tmpFile).Decode(loadedMVC)
	assert.Nil(t, err)

	if !assert.ObjectsAreEqual(*mvc, *loadedMVC) {
		t.Errorf("expected %+v, got %+v", *mvc, *loadedMVC)
	}
}

func TestHandleAllConfigSetup(t *testing.T) {
	t.Run("config-file-does-exist", func(t *testing.T) {
		const sampleConfig = `
pid = "/pid"
sleep = 1.0
ignore_ssl_errors = true
`

		tmpFile, err := ioutil.TempFile("", "")
		assert.Nil(t, err)
		defer os.Remove(tmpFile.Name())

		err = ioutil.WriteFile(tmpFile.Name(), []byte(sampleConfig), 0755)
		assert.Nil(t, err)

		config, err := HandleAllConfigSetup(tmpFile.Name())
		assert.Nil(t, err)

		assert.Equal(t, "/pid", config.PidFile)
		assert.Equal(t, 1.0, config.Sleep)
		assert.Equal(t, true, config.IgnoreSSLErrors)
	})

	t.Run("config-file-does-not-exist", func(t *testing.T) {
		// Create a temp file to get a file path we can use for temp
		// config generation. But delete it so we can acutally write our
		// config file under the path.
		tmpFile, err := ioutil.TempFile("", "")
		assert.Nil(t, err)
		configFilePath := tmpFile.Name()
		err = os.Remove(tmpFile.Name())
		assert.Nil(t, err)

		_, err = HandleAllConfigSetup(configFilePath)
		assert.Nil(t, err)

		_, err = os.Stat(configFilePath)
		assert.Nil(t, err)

		mvc := NewMinimumConfig()
		loadedMVC := &MinValuableConfig{}
		err = toml.NewDecoder(tmpFile).Decode(loadedMVC)
		assert.Nil(t, err)

		if !assert.ObjectsAreEqual(*mvc, *loadedMVC) {
			t.Errorf("expected %+v, got %+v", *mvc, *loadedMVC)
		}
	})
}
