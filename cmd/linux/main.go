/*
TapTo
Copyright (C) 2023 Gareth Jones
Copyright (C) 2023, 2024 Callan Barrett

This file is part of TapTo.

TapTo is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

TapTo is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with TapTo.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/wizzomafizzo/tapto/pkg/launcher"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
	"github.com/wizzomafizzo/tapto/pkg/utils"

	"github.com/wizzomafizzo/mrext/pkg/input"

	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon"
)

const (
	appName    = "tapto"
	appVersion = "1.4"
)

func main() {
	svcOpt := flag.String("service", "", "manage TapTo service (start, stop, restart, status)")
	launchOpt := flag.String("launch", "", "execute given text as if it were a token")
	flag.Parse()

	err := utils.InitLogging()
	if err != nil {
		fmt.Println("Error initializing logging:", err)
		os.Exit(1)
	}

	cfg, err := config.NewUserConfig(appName, &config.UserConfig{
		TapTo: config.TapToConfig{
			ProbeDevice: true,
		},
	})
	if err != nil {
		log.Error().Msgf("error loading user config: %s", err)
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	log.Info().Msgf("TapTo v%s", appVersion)
	log.Info().Msgf("config path = %s", cfg.IniPath)
	log.Info().Msgf("app path = %s", cfg.AppPath)
	log.Info().Msgf("connection_string = %s", cfg.GetConnectionString())
	log.Info().Msgf("allow_commands = %t", cfg.GetAllowCommands())
	log.Info().Msgf("disable_sounds = %t", cfg.GetDisableSounds())
	log.Info().Msgf("probe_device = %t", cfg.GetProbeDevice())
	log.Info().Msgf("exit_game = %t", cfg.GetExitGame())
	log.Info().Msgf("exit_game_blocklist = %s", cfg.GetExitGameBlocklist())
	log.Info().Msgf("debug = %t", cfg.GetDebug())

	if cfg.GetDebug() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	svc, err := mister.NewService(mister.ServiceArgs{
		Name: appName,
		Entry: func() (func() error, error) {
			return daemon.StartDaemon(cfg)
		},
	})
	if err != nil {
		log.Error().Msgf("error creating service: %s", err)
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}

	if *launchOpt != "" {
		kbd, err := input.NewKeyboard()
		if err != nil {
			log.Error().Msgf("error creating keyboard: %s", err)
			fmt.Println("Error creating keyboard:", err)
			os.Exit(1)
		}

		// TODO: this is doubling up on the split logic in daemon
		cmds := strings.Split(*launchOpt, "||")
		for i, cmd := range cmds {
			err, _ := launcher.LaunchToken(cfg, true, kbd, cmd, len(cmds), i)
			if err != nil {
				log.Error().Msgf("error launching token: %s", err)
				fmt.Println("Error launching token:", err)
				os.Exit(1)
			}
		}

		os.Exit(0)
	}

	svc.ServiceHandler(svcOpt)
}
