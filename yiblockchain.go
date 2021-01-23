package main

import (
	"flag"
	"fmt"
	"github.com/jacoblai/yiblockchain/store"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	tmlog "github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "$HOME/.tendermint/config/config.toml", "Path to config.toml")
	flag.Parse()
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	app, err := store.NewLvStore(dir)
	if err != nil {
		log.Fatal(err)
	}

	if strings.Contains(configFile, "$HOME") {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		configFile = strings.Replace(configFile, "$HOME", usr.HomeDir, 1)
	}

	// read config
	config := cfg.DefaultConfig()
	config.RootDir = filepath.Dir(filepath.Dir(configFile))
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err, "viper failed to read config file")
	}
	if err := viper.Unmarshal(config); err != nil {
		log.Fatal(err, "viper failed to unmarshal config")
	}
	if err := config.ValidateBasic(); err != nil {
		log.Fatal(err, "config is invalid")
	}
	// create logger
	logger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout))
	logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel())
	if err != nil {
		log.Fatal(err, "failed to parse log level")
	}
	// read private validator
	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)
	// read node key
	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		log.Fatal(err, "failed to load node's key")
	}
	// create node
	node, err := nm.NewNode(
		config,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(config),
		nm.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger)
	if err != nil {
		log.Fatal(err, "failed to create new Tendermint node")
	}

	err = node.Start()
	if err != nil {
		log.Fatal(err, "failed to run new Tendermint node")
	}

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	cleanup := make(chan bool)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range signalChan {
			err = node.Stop()
			if err != nil {
				cleanup <- true
			}
			go func() {
				node.Wait()
				cleanup <- true
			}()
			<-cleanup
			fmt.Println("safe exit")
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
