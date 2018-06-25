package create

import (
    "os"
    "path/filepath"
    "strings"
    "wio/cmd/wio/config"
    "wio/cmd/wio/errors"
    "wio/cmd/wio/log"
    "wio/cmd/wio/utils"
    "wio/cmd/wio/utils/io"
    "wio/cmd/wio/utils/template"
)

func (info createInfo) fillReadMe(queue *log.Queue, readmeFile string) error {
    log.Verb(queue, "filling README file ... ")
    if err := template.IOReplace(readmeFile, map[string]string{
        "PLATFORM":        strings.Join(info.PlatformNamesDetailed, ","),
        "BOARD":           strings.Join(info.BoardNamesDetailed, ","),
        "PROJECT_NAME":    info.Name,
        "PROJECT_VERSION": "0.0.1",
    }); err != nil {
        log.WriteFailure(queue, log.VERB)
        return err
    }
    log.WriteSuccess(queue, log.VERB)
    return nil
}

func (info createInfo) toLowerCase() {
    info.Type = strings.ToLower(info.Type)
    info.PlatformNamesSimple = strings.Split(strings.ToLower(strings.Join(info.PlatformNamesSimple, "{}")), "{}")
    info.PlatformNamesDetailed = strings.Split(strings.ToLower(strings.Join(info.PlatformNamesDetailed, "{}")), "{}")
    info.BoardNamesSimple = strings.Split(strings.ToLower(strings.Join(info.BoardNamesSimple, "{}")), "}")
    info.BoardNamesDetailed = strings.Split(strings.ToLower(strings.Join(info.BoardNamesDetailed, "{}")), "{}")
}

func (create Create) generateConstraints() (map[string]bool, map[string]bool) {
    context := create.Context
    dirConstraints := map[string]bool{
        "tests":          false,
        "no-header-only": !context.Bool("header-only"),
    }
    fileConstraints := map[string]bool{
        "ide=clion":      false,
        "extra":          !context.Bool("no-extras"),
        "example":        context.Bool("create-example"),
        "no-header-only": !context.Bool("no-header-only"),
    }
    return dirConstraints, fileConstraints
}

// This uses a structure.json file and creates a project structure based on that. It takes in consideration
// all the constrains and copies files. This should be used for creating project for any type of app/pkg
func (create Create) copyProjectAssets(queue *log.Queue, info *createInfo, data StructureTypeData) error {
    dirConstraints, fileConstraints := create.generateConstraints()
    for _, path := range data.Paths {
        directoryPath := filepath.Clean(info.Directory + io.Sep + path.Entry)
        skipDir := false
        log.Verbln(queue, "copying assets to directory: %s", directoryPath)
        // handle directory constraints
        for _, constraint := range path.Constraints {
            _, exists := dirConstraints[constraint]
            if exists && !dirConstraints[constraint] {
                log.Verbln(queue, "constraint not specified and hence skipping this directory")
                skipDir = true
                break
            }
        }
        if skipDir {
            continue
        }

        if !utils.PathExists(directoryPath) {
            if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
                return err
            }
            log.Verbln(queue, "created directory: %s", directoryPath)
        }

        log.Verbln(queue, "copying asset files for directory: %s", directoryPath)
        for _, file := range path.Files {
            toPath := filepath.Clean(directoryPath + io.Sep + file.To)
            skipFile := false
            // handle file constraints
            for _, constraint := range file.Constraints {
                _, exists := fileConstraints[constraint]
                if exists && !fileConstraints[constraint] {
                    log.Verbln(queue, "constraint not specified and hence skipping this file")
                    skipFile = true
                    break
                }
            }
            if skipFile {
                continue
            }

            // handle updates
            if !file.Update && create.Update {
                log.Verbln(queue, "project is not updating, hence skipping update for path: %s", toPath)
                continue
            }

            // copy assets
            if err := io.AssetIO.CopyFile(file.From, toPath, file.Override); err != nil {
                return err
            } else {
                log.Verbln(queue, `copied asset file "%s" TO: %s: `, filepath.Base(file.From), toPath)
            }
        }
    }
    return nil
}

// This parses platform and board structure for project of pkg type
func (create Create) parsePlatformsAndBoards(platformsString string, boardsString string) (
    map[string]*PlatformInfo, []string, []string, []string, []string) {

    platformsGiven := strings.Split(platformsString, " ")

    platforms := map[string]*PlatformInfo{}
    var platformNamesSimple []string
    var platformNamesDetailed []string
    var boardNamesSimple []string
    var boardNamesDetailed []string

    for _, platformGiven := range platformsGiven {
        given := strings.Split(platformGiven, ":")

        var platform *PlatformInfo

        if val, exists := platforms[given[0]]; exists {
            platform = val
        } else {
            platforms[given[0]] = &PlatformInfo{}
            platform = platforms[given[0]]
            platform.Frameworks = []string{}
            platform.Boards = []string{}
        }

        platformNamesSimple = append(platformNamesSimple, given[0])

        if len(given) > 1 {
            platform.Frameworks = append(platform.Frameworks, given[1])
            platformNamesDetailed = append(platformNamesDetailed, given[0]+":"+given[1])
        } else {
            // choose a default framework
            switch given[0] {
            case "all":
                platform.Frameworks = append(platform.Frameworks, "all")
                platformNamesDetailed = append(platformNamesDetailed, given[0]+":all")
            case "atmelavr", "native":
                platform.Frameworks = append(platform.Frameworks, config.FrameworkDefaults[given[0]])
                platformNamesDetailed = append(platformNamesDetailed, given[0]+":"+config.FrameworkDefaults[given[0]])
            default:
                log.WriteErrorlnExit(errors.PlatformNotSupportedError{
                    Platform: given[0],
                })
            }
        }
    }

    // make all boards supported for all platforms
    if boardsString == "all" {
        boardsString = ""
        for _, platformName := range platformNamesSimple {
            boardsString += platformName + ":all"
        }
    }

    boardsGiven := strings.Split(boardsString, " ")

    for _, boardGiven := range boardsGiven {
        given := strings.Split(boardGiven, ":")

        var platform *PlatformInfo

        if val, exists := platforms[given[0]]; exists {
            platform = val
        } else {
            continue
        }

        switch given[0] {
        case "atmelavr":
            if len(given) > 1 {
                platform.Boards = append(platform.Boards, given[1])
                boardNamesSimple = append(boardNamesSimple, given[0])
                boardNamesDetailed = append(boardNamesDetailed, given[0]+":"+given[1])
            } else {
                // choose a default framework
                platform.Boards = append(platform.Boards, config.BoardDefaults[given[0]])
                boardNamesSimple = append(boardNamesSimple, given[0])
                boardNamesDetailed = append(boardNamesDetailed, given[0]+":"+config.BoardDefaults[given[0]])
            }
        }
    }

    return platforms, platformNamesSimple, platformNamesDetailed, boardNamesSimple, boardNamesDetailed
}
