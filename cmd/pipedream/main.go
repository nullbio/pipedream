package main

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/aarondl/zapcolors"
	"github.com/davecgh/go-spew/spew"
	"github.com/nullbio/pipedream"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/uber-go/zap"
)

const (
	envPrefix = "PIPEDREAM_"
)

var rootCmd = cobra.Command{
	Use:              "pipedream [flags]",
	Short:            "Pipedream precompiles your assets",
	PersistentPreRun: setup,
	Run:              rootCmdCobra,
}

var (
	log zap.Logger

	pipeline pipedream.Pipedream

	flagNoColor bool
	flagConfig  string

	flagIn         string
	flagOut        string
	flagCDNURL     string
	flagNoCompile  bool
	flagNoMinify   bool
	flagNoHash     bool
	flagNoCompress bool
)

func main() {
	flags := rootCmd.PersistentFlags()
	flags.BoolVarP(&flagNoColor, "no-color", "", false, "No color output")
	flags.StringVarP(&flagConfig, "config", "c", "", "Path to a configuration file")

	flags.StringVarP(&flagIn, "in", "i", "", "The input directory (usually project/assets)")
	flags.StringVarP(&flagOut, "out", "o", "", "The output directory (usually project/compiled)")

	flags.StringVarP(&flagCDNURL, "cdn-url", "", "", "An optional CDN URL")

	flags.BoolVarP(&flagNoCompile, "no-compile", "", false, "Don't run assets through compilers")
	flags.BoolVarP(&flagNoMinify, "no-minify", "", false, "Don't run assets through minifiers")
	flags.BoolVarP(&flagNoHash, "no-hash", "", false, "Don't fingerprint the end result, also disables manifest generation")
	flags.BoolVarP(&flagNoCompress, "no-compress", "", false, "Don't generate .gz copies of the files")

	if err := rootCmd.Execute(); err != nil {
		if err != nil {
			os.Exit(1)
		}
	}
}

func setup(cmd *cobra.Command, args []string) {
	if flagNoColor {
		log = zap.New(zap.NewTextEncoder())
	} else {
		log = zap.New(zapcolors.NewColorEncoder())
	}

	setConfigString(&flagConfig, "config")
	setConfigBool(&flagNoColor, "no-color")

	log.Info("reading config", zap.String("file", flagConfig))

	if _, err := toml.DecodeFile(flagConfig, &pipeline); err != nil {
		log.Fatal("failed to read config", zap.Error(err))
	}

	val := reflect.ValueOf(&pipeline)
	typ := val.Elem().Type()
	val = reflect.Indirect(val)

	n := typ.NumField()
	for i := 0; i < n; i++ {
		st := typ.Field(i)
		tag := st.Tag.Get("toml")

		if st.Anonymous || tag == "-" {
			continue
		}

		kind := st.Type.Kind()
		switch kind {
		case reflect.String:
			setConfigString(val.Field(i).Addr().Interface().(*string), tag)
		case reflect.Bool:
			setConfigBool(val.Field(i).Addr().Interface().(*bool), tag)
		default:
			continue
		}
	}
}

func rootCmdCobra(cmd *cobra.Command, args []string) {
	spew.Dump(pipeline)
}

func setConfigString(inStruct *string, name string) {
	flag := pflag.Lookup(name)
	if flag != nil && flag.Changed {
		*inStruct = flag.Value.String()
		return
	}

	if env := tagEnv(name); len(env) != 0 {
		*inStruct = env
	}
}

func setConfigBool(inStruct *bool, name string) {
	var strval string
	if flag := pflag.Lookup(name); flag != nil && flag.Changed {
		strval = flag.Value.String()
	} else if env := tagEnv(name); len(env) != 0 {
		strval = env
	} else {
		return
	}

	var err error
	if *inStruct, err = strconv.ParseBool(strval); err != nil {
		log.Fatal("failed to parse bool", zap.String("config-key", name))
	}
}

func tagEnv(name string) string {
	return os.Getenv(envPrefix + strings.ToUpper(name))
}
