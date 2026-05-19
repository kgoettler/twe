/*
Copyright © 2024 Ken Goettler <goettlek@gmail.com>
*/
//nolint: gochecknoglobals, gochecknoinits // not applicable to cobra-cli files
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var ConfigKeys = map[string]struct{}{}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Read and write twe configuration",
}

func RunCmdConfigGet(args []string) (any, error) {
	if len(args) == 0 {
		b, err := readConfigFile()
		if err != nil {
			return nil, fmt.Errorf("getting config: %w", err)
		}
		bstr := string(b)
		return bstr, nil
	}
	key := args[0]
	if _, ok := ConfigKeys[key]; !ok {
		return nil, fmt.Errorf("invalid config key: '%s'", key)
	}

	val := viper.Get(key)
	return val, nil
}

func RunCmdConfigSet(cfgPath string, key string, value string) error {
	// Read existing config; tolerate a missing file.
	existing := make(map[string]interface{})
	raw, err := os.ReadFile(cfgPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading config file: %s", err)
	}
	if err == nil {
		if err = yaml.Unmarshal(raw, &existing); err != nil {
			return fmt.Errorf("parsing config file: %s", err)
		}
	}
	if _, ok := ConfigKeys[key]; !ok {
		return fmt.Errorf("invalid config key: '%s'", key)
	}
	setNestedKey(existing, key, parseScalar(value))

	if err = os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %s", err)
	}
	out, err := yaml.Marshal(existing)
	if err != nil {
		return fmt.Errorf("encoding config: %s", err)
	}
	if err = os.WriteFile(cfgPath, out, 0o644); err != nil { //nolint: gosec // config file, not sensitive
		return fmt.Errorf("writing config file: %s", err)
	}
	return nil
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Print the value of a config key",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		val, err := RunCmdConfigGet(args)
		if err != nil {
			handleError(cmd, "getting config: %s", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), val)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key, rawVal := args[0], args[1]
		cfgPath := resolveConfigFilePath()

		err := RunCmdConfigSet(cfgPath, key, rawVal)
		if err != nil {
			handleError(cmd, "setting config: %s", err)
		}
	},
}

func readConfigFile() ([]byte, error) {
	filepath := resolveConfigFilePath()
	raw, err := os.ReadFile(filepath)
	if err != nil && os.IsNotExist(err) {
		return raw, nil
	} else if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	return raw, nil
}

// parseScalar converts a raw string to int, bool, or string (in that order).
func parseScalar(s string) interface{} {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return s
}

// setNestedKey sets a dot-separated key path in a nested map.
func setNestedKey(m map[string]interface{}, key string, value interface{}) {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) == 1 {
		m[key] = value
		return
	}
	sub, _ := m[parts[0]].(map[string]interface{})
	if sub == nil {
		sub = make(map[string]interface{})
	}
	setNestedKey(sub, parts[1], value)
	m[parts[0]] = sub
}

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}
