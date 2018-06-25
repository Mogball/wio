// Copyright 2018 Waterloop. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Part of commands/create package, which contains create and update command and sub commands provided by the tool.
// Creates, updates and initializes a wio project.
package create

import (
    "github.com/fatih/color"
    "path/filepath"
    "strings"
    "wio/cmd/wio/config"
    "wio/cmd/wio/constants"
    "wio/cmd/wio/log"
    "wio/cmd/wio/types"
    "wio/cmd/wio/utils/io"
)

// Creation of AVR projects
func (create Create) createPackageProject(dir string) {
    platforms, platformNamesSimple, platformNamesDetailed, boardNamesSimple, boardNamesDetailed :=
        create.parsePlatformsAndBoards(create.Context.String("platform"), create.Context.String("board"))

    info := createInfo{
        Directory:             dir,
        Type:                  string(constants.PKG),
        Name:                  filepath.Base(dir),
        Platform:              platforms,
        PlatformNamesSimple:   platformNamesSimple,
        PlatformNamesDetailed: platformNamesDetailed,
        BoardNamesSimple:      boardNamesSimple,
        BoardNamesDetailed:    boardNamesDetailed,
        ConfigOnly:            create.Context.Bool("only-config"),
        HeaderOnly:            create.Context.Bool("header-only"),
    }
    info.toLowerCase()

    // Generate project structure
    queue := log.GetQueue()
    if !info.ConfigOnly {
        log.Info(log.Cyan, "creating project structure ... ")
        if err := create.createPackageStructure(queue, &info); err != nil {
            log.WriteFailure()
            log.WriteErrorlnExit(err)
        } else {
            log.WriteSuccess()
        }
        log.PrintQueue(queue, log.TWO_SPACES)
    }

    // Fill configuration file
    queue = log.GetQueue()
    log.Info(log.Cyan, "configuring project files ... ")
    if err := create.fillPackageConfig(queue, &info); err != nil {
        log.WriteFailure()
        log.WriteErrorlnExit(err)
    } else {
        log.WriteSuccess()
    }
    log.PrintQueue(queue, log.TWO_SPACES)

    // print structure summary
    info.printPackageCreateSummary()
}

// Copy and generate files for a package project
func (create Create) createPackageStructure(queue *log.Queue, info *createInfo) error {
    log.Verb(queue, "reading paths.json file ... ")
    structureData := &StructureConfigData{}

    // read configurationsFile
    if err := io.AssetIO.ParseJson("configurations/structure-atmelavr.json", structureData); err != nil {
        log.WriteFailure(queue, log.VERB)
        return err
    } else {
        log.WriteSuccess(queue, log.VERB)
    }

    log.Verb(queue, "copying asset files ... ")
    subQueue := log.GetQueue()

    if err := create.copyProjectAssets(subQueue, info, structureData.Pkg); err != nil {
        log.WriteFailure(queue, log.VERB)
        log.CopyQueue(subQueue, queue, log.FOUR_SPACES)
        return err
    } else {
        log.WriteSuccess(queue, log.VERB)
        log.CopyQueue(subQueue, queue, log.FOUR_SPACES)
    }

    readmeFile := info.Directory + io.Sep + "README.md"
    err := info.fillReadMe(queue, readmeFile)

    return err
}

// Generate wio.yml for package project
func (create Create) fillPackageConfig(queue *log.Queue, info *createInfo) error {
    /*// handle app
      if create.Type == constants.APP {
          log.QueueWrite(queue, log.INFO, nil, "creating config file for application ... ")

          appConfig := &types.AppConfig{}
          appConfig.MainTag.Name = filepath.Base(directory)
          appConfig.MainTag.Ide = config.ProjectDefaults.Ide

          // supported board, framework and platform and wio version
          fillMainTagConfiguration(&appConfig.MainTag.Config, []string{board}, constants.AVR, []string{framework})

          appConfig.MainTag.CompileOptions.Platform = constants.AVR

          // create app target
          appConfig.TargetsTag.DefaultTarget = config.ProjectDefaults.AppTargetName
          appConfig.TargetsTag.Targets = map[string]types.AppAVRTarget{
              config.ProjectDefaults.AppTargetName: {
                  Src:       "src",
                  Framework: framework,
                  Board:     board,
                  Flags: types.AppTargetFlags{
                      GlobalFlags: []string{},
                      TargetFlags: []string{},
                  },
              },
          }

          projectConfig = appConfig
      } else {*/
    log.Verb(queue, "creating config file for package ... ")

    visibility := "PRIVATE"
    if info.HeaderOnly {
        visibility = "INTERFACE"
    }

    var targets types.PkgAVRTargets

    if _, exists := info.Platform["all"]; !exists {
        var targetPlatform string
        var targetFramework string
        var targetBoard string

        // handle boards case where it is all or none
        targetPlatform = info.PlatformNamesSimple[0]
        targetFramework = info.Platform[targetPlatform].Frameworks[0]

        if len(info.Platform[targetPlatform].Boards) == 0 {
            targetBoard = "null"
        } else {
            if info.Platform[targetPlatform].Boards[0] == "all" {
                targetBoard = config.BoardDefaults[targetPlatform]
            } else {
                targetBoard = info.Platform[targetPlatform].Boards[0]
            }
        }

        target := config.DefaultTargetDefaults
        targets = types.PkgAVRTargets{
            DefaultTarget: config.DefaultTargetDefaults,
            Targets: map[string]types.PkgAVRTarget{
                target: {
                    Src:       config.TargetSourceDefault[constants.PKG],
                    Platform:  targetPlatform,
                    Framework: targetFramework,
                    Board:     targetBoard,
                },
            },
        }
    } else {
        // create an empty target
        targets = types.PkgAVRTargets{
            DefaultTarget: "null",
            Targets:       nil,
        }
    }

    projectConfig := &types.PkgConfig{
        MainTag: types.PkgTag{
            Ide: config.IdeDefault,
            Meta: types.PackageMeta{
                Name:    info.Name,
                Version: "0.0.1",
                License: "MIT",
            },
            Config: types.Configurations{
                WioVersion:         config.ProjectMeta.Version,
                HeaderOnly:         info.HeaderOnly,
                SupportedPlatforms: info.PlatformNamesDetailed,
                SupportedBoards:    info.BoardNamesDetailed,
            },
            Flags:       types.Flags{Visibility: visibility},
            Definitions: types.Definitions{Visibility: visibility},
        },
        TargetsTag: targets,
    }

    log.WriteSuccess(queue, log.VERB)
    log.Verb(queue, "pretty printing wio.yml file ... ")
    wioYmlPath := info.Directory + io.Sep + "wio.yml"
    if err := projectConfig.PrettyPrint(wioYmlPath, false); err != nil {
        log.WriteFailure(queue, log.VERB)
        return err
    }
    log.WriteSuccess(queue, log.VERB)
    return nil
}

// Print package creation summary
func (info createInfo) printPackageCreateSummary() {
    log.Writeln()
    log.Infoln(log.Yellow.Add(color.Underline), "Project structure summary")
    if !info.HeaderOnly {
        log.Info(log.Cyan, "src              ")
        log.Writeln("source/non client files")
    }
    log.Info(log.Cyan, "tests            ")
    log.Writeln("source files for test target")
    log.Info(log.Cyan, "include          ")
    log.Writeln("public headers for the package")

    // print project summary
    log.Writeln()
    log.Infoln(log.Yellow.Add(color.Underline), "Project creation summary")
    log.Info(log.Cyan, "path             ")
    log.Writeln(info.Directory)
    log.Info(log.Cyan, "project type     ")
    log.Writeln("pkg")
    log.Info(log.Cyan, "platform         ")
    log.Writeln(strings.Join(info.PlatformNamesDetailed, ","))
    log.Info(log.Cyan, "board            ")
    log.Writeln(strings.Join(info.BoardNamesDetailed, ","))
}
