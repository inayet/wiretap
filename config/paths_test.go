// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pb33f/wiretap/shared"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func TestFindPath(t *testing.T) {

	config := `
paths:
  /pb33f/test/**:
    target: /
    secure: false
    pathRewrite:
      '^/pb33f/test/': ''`

	viper.SetConfigType("yaml")
	verr := viper.ReadConfig(strings.NewReader(config))
	assert.NoError(t, verr)

	paths := viper.Get("paths")
	var pc map[string]*shared.WiretapPathConfig

	derr := mapstructure.Decode(paths, &pc)
	assert.NoError(t, derr)

	wcConfig := &shared.WiretapConfiguration{
		PathConfigurations: pc,
	}

	wcConfig.CompilePaths()

	res := FindPaths("/pb33f/test/123", wcConfig)
	assert.Len(t, res, 1)

	res = FindPaths("/pb33f/test/123/sing/song", wcConfig)
	assert.Len(t, res, 1)

	res = FindPaths("/pb33f/no-match/wrong", wcConfig)
	assert.Len(t, res, 0)

}

func TestRewritePath(t *testing.T) {

	config := `
paths:
  /pb33f/test/**:
    target: localhost:9093/
    secure: false
    pathRewrite:
      '^/pb33f/test/': ''`

	viper.SetConfigType("yaml")
	verr := viper.ReadConfig(strings.NewReader(config))
	assert.NoError(t, verr)

	paths := viper.Get("paths")
	var pc map[string]*shared.WiretapPathConfig

	derr := mapstructure.Decode(paths, &pc)
	assert.NoError(t, derr)

	wcConfig := &shared.WiretapConfiguration{
		PathConfigurations: pc,
	}

	wcConfig.CompilePaths()

	path := RewritePath("/pb33f/test/123/slap/a/chap", wcConfig)
	assert.Equal(t, "http://localhost:9093/123/slap/a/chap", path)

}

func TestRewritePath_Secure(t *testing.T) {

	config := `
paths:
  /pb33f/*/test/**:
    target: localhost:9093
    secure: true
    pathRewrite:
      '^/pb33f/(\w+)/test/': '/flat/jam/'`

	viper.SetConfigType("yaml")
	verr := viper.ReadConfig(strings.NewReader(config))
	assert.NoError(t, verr)

	paths := viper.Get("paths")
	var pc map[string]*shared.WiretapPathConfig

	derr := mapstructure.Decode(paths, &pc)
	assert.NoError(t, derr)

	wcConfig := &shared.WiretapConfiguration{
		PathConfigurations: pc,
	}

	wcConfig.CompilePaths()

	path := RewritePath("/pb33f/cakes/test/123/smelly/jelly", wcConfig)
	assert.Equal(t, "https://localhost:9093/flat/jam/123/smelly/jelly", path)

}

func TestRewritePath_Secure_With_Variables(t *testing.T) {

	config := `
paths:
  /pb33f/*/test/*/321/**:
    target: localhost:9093
    secure: true
    pathRewrite:
      '^/pb33f/(\w+)/test/(\w+)/(\d+)/': '/slippy/$1/whip/$3/$2/'`

	viper.SetConfigType("yaml")
	verr := viper.ReadConfig(strings.NewReader(config))
	assert.NoError(t, verr)

	paths := viper.Get("paths")
	var pc map[string]*shared.WiretapPathConfig

	derr := mapstructure.Decode(paths, &pc)
	assert.NoError(t, derr)

	wcConfig := &shared.WiretapConfiguration{
		PathConfigurations: pc,
	}

	wcConfig.CompilePaths()

	path := RewritePath("/pb33f/cakes/test/lemons/321/smelly/jelly", wcConfig)
	assert.Equal(t, "https://localhost:9093/slippy/cakes/whip/321/lemons/smelly/jelly", path)

}

func TestRewritePath_Secure_With_Variables_CaseSensitive(t *testing.T) {

	config := `
paths:
  /en-US/burgerd/__raw/*:
    target: localhost:80
    pathRewrite:
      '^/en-US/burgerd/__raw/(\w+)/nobody/': '$1/-/'
  /en-US/burgerd/services/*:
    target: locahost:80
    pathRewrite:
      '^/en-US/burgerd/services': '/services'`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompilePaths()

	path := RewritePath("/en-US/burgerd/__raw/noKetchupPlease/nobody/", &c)
	assert.Equal(t, "http://localhost:80/noKetchupPlease/-/", path)

}

func TestRewritePath_Secure_With_Variables_CaseSensitive_AndQuery(t *testing.T) {

	config := `
paths:
  /en-US/burgerd/__raw/*:
    target: localhost:80
    pathRewrite:
      '^/en-US/burgerd/__raw/(\w+)/nobody/': '$1/-/'
  /en-US/burgerd/services/*:
    target: locahost:80
    pathRewrite:
      '^/en-US/burgerd/services': '/services'`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompilePaths()

	path := RewritePath("/en-US/burgerd/__raw/noKetchupPlease/nobody/yummy/yum?onions=true", &c)
	assert.Equal(t, "http://localhost:80/noKetchupPlease/-/yummy/yum?onions=true", path)

}

func TestLocatePathDelay(t *testing.T) {

	config := `pathDelays:
  /pb33f/test/**: 1000
  /pb33f/cakes/123: 2000
  /*/test/123: 3000`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompilePathDelays()

	delay := FindPathDelay("/pb33f/test/burgers/fries?1234=no", &c)
	assert.Equal(t, 1000, delay)

	delay = FindPathDelay("/pb33f/cakes/123", &c)
	assert.Equal(t, 2000, delay)

	delay = FindPathDelay("/roastbeef/test/123", &c)
	assert.Equal(t, 3000, delay)

	delay = FindPathDelay("/not-registered", &c)
	assert.Equal(t, 0, delay)

}
