// Copyright 2020 Buf Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build darwin linux

package app

import (
	"errors"
	"fmt"
	"path/filepath"
)

// DevNullFilePath is the equivalent of /dev/null.
//
// This will be /dev/null for darwin and linux.
// This will be nul for windows.
const DevNullFilePath = "/dev/null"

// HomeDirPath returns the home directory path.
//
// This will be $HOME for darwin and linux.
// This will be %USERPROFILE% for windows.
func HomeDirPath(envContainer EnvContainer) (string, error) {
	if home := envContainer.Env("HOME"); home != "" {
		return home, nil
	}
	return "", errors.New("$HOME is not set")
}

// CacheDirPath returns the cache directory path.
//
// This will be $XDG_CACHE_HOME for darwin and linux, falling back to $HOME/.cache.
// This will be %LocalAppData% for windows.
//
// Users cannot assume that CacheDirPath, ConfigDirPath, and DataDirPath are unique.
func CacheDirPath(envContainer EnvContainer) (string, error) {
	return xdgDirPath(envContainer, "XDG_CACHE_HOME", ".cache")
}

// ConfigDirPath returns the config directory path.
//
// This will be $XDG_CONFIG_HOME for darwin and linux, falling back to $HOME/.config.
// This will be %AppData% for windows.
//
// Users cannot assume that CacheDirPath, ConfigDirPath, and DataDirPath are unique.
func ConfigDirPath(envContainer EnvContainer) (string, error) {
	return xdgDirPath(envContainer, "XDG_CONFIG_HOME", ".config")
}

// DataDirPath returns the data directory path.
//
// This will be $XDG_DATA_HOME for darwin and linux, falling back to $HOME/.local/share.
// This will be %LocalAppData% for windows.
//
// Users cannot assume that CacheDirPath, ConfigDirPath, and DataDirPath are unique.
func DataDirPath(envContainer EnvContainer) (string, error) {
	return xdgDirPath(envContainer, "XDG_DATA_HOME", filepath.Join(".local", "share"))
}

func xdgDirPath(envContainer EnvContainer, key string, fallbackRelHomeDirPath string) (string, error) {
	if value := envContainer.Env(key); value != "" {
		return value, nil
	}
	if home := envContainer.Env("HOME"); home != "" {
		return filepath.Join(home, fallbackRelHomeDirPath), nil
	}
	return "", fmt.Errorf("$%s and $HOME are not set", key)
}
