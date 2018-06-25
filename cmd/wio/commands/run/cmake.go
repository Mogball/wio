package run

import (
    "os"
    "path/filepath"
    "strings"
    "wio/cmd/wio/constants"
    "wio/cmd/wio/errors"
    "wio/cmd/wio/utils"
    "wio/cmd/wio/utils/io"
)

// CMake Target information
type CMakeTarget struct {
    TargetName            string
    Path                  string
    Flags                 []string
    Definitions           []string
    FlagsVisibility       string
    DefinitionsVisibility string
    HeaderOnly            bool
}

// CMake Target Link information
type CMakeTargetLink struct {
    From           string
    To             string
    LinkVisibility string
}

// This creates CMake library string that will be used to link libraries
func GenerateAtmelAvrDependencyCMakeString(targets map[string]*CMakeTarget, links []CMakeTargetLink) []string {
    cmakeStrings := make([]string, 0)

    for _, target := range targets {
        finalString := atmelavrNonHeaderOnlyString

        if target.HeaderOnly {
            finalString = atmelavrHeaderOnlyString
        }

        finalString = strings.Replace(finalString, "{{DEPENDENCY_NAME}}", target.TargetName, -1)
        finalString = strings.Replace(finalString, "{{DEPENDENCY_PATH}}", filepath.ToSlash(target.Path), -1)
        finalString = strings.Replace(finalString, "{{DEPENDENCY_FLAGS}}",
            strings.Join(target.Flags, " "), -1)
        finalString = strings.Replace(finalString, "{{DEPENDENCY_DEFINITIONS}}",
            strings.Join(target.Definitions, " "), -1)
        finalString = strings.Replace(finalString, "{{FLAGS_VISIBILITY}}", target.FlagsVisibility, -1)
        finalString = strings.Replace(finalString, "{{DEFINITIONS_VISIBILITY}}", target.DefinitionsVisibility, -1)

        cmakeStrings = append(cmakeStrings, finalString+"\n")
    }

    for _, link := range links {
        finalString := linkString
        finalString = strings.Replace(finalString, "{{LINKER_NAME}}", link.From, -1)
        finalString = strings.Replace(finalString, "{{DEPENDENCY_NAME}}", link.To, -1)

        finalString = strings.Replace(finalString, "{{LINK_VISIBILITY}}", link.LinkVisibility, -1)

        cmakeStrings = append(cmakeStrings, finalString)
    }

    cmakeStrings = append(cmakeStrings, "")

    return cmakeStrings
}

// THis Creates the main CMakeLists.txt file for AVR app type project
func GenerateAtmelAvrMainCMakeLists(appName string, appPath string, targetInfo TargetBuildInfo) error {

    executablePath, err := io.NormalIO.GetRoot()
    if err != nil {
        return err
    }

    var toolChainPath string
    if targetInfo.Framework == constants.COSA {
        toolChainPath = "toolchain/cmake/CosaToolchain.cmake"
    } else {
        return errors.FrameworkNotSupportedError{
            Platform:  constants.ATMELAVR,
            Framework: targetInfo.Framework,
        }
    }

    // read the CMakeLists.txt file template
    templateData, err := io.AssetIO.ReadFile("templates/cmake/CMakeListsAtmelAVR.txt.tpl")
    if err != nil {
        return err
    }

    templateDataStr := strings.Replace(string(templateData), "{{TOOLCHAIN_PATH}}",
        filepath.ToSlash(executablePath), -1)
    templateDataStr = strings.Replace(templateDataStr, "{{TOOLCHAIN_FILE_REL}}",
        filepath.ToSlash(toolChainPath), -1)
    templateDataStr = strings.Replace(templateDataStr, "{{PROJECT_PATH}}", filepath.ToSlash(appPath), -1)
    templateDataStr = strings.Replace(templateDataStr, "{{PROJECT_NAME}}", appName, -1)
    templateDataStr = strings.Replace(templateDataStr, "{{TARGET_NAME}}", targetInfo.TargetName, -1)
    templateDataStr = strings.Replace(templateDataStr, "{{BOARD}}", targetInfo.Board, -1)
    templateDataStr = strings.Replace(templateDataStr, "{{PORT}}", targetInfo.Port, -1)
    templateDataStr = strings.Replace(templateDataStr, "{{FRAMEWORK}}", strings.Title(targetInfo.Framework), -1)
    templateDataStr = strings.Replace(templateDataStr, "{{ENTRY}}", targetInfo.Src, 1)
    templateDataStr = strings.Replace(templateDataStr, "{{TARGET_COMPILE_FLAGS}}",
        strings.Join(targetInfo.Flags.GetTargetFlags(), " "), -1)
    templateDataStr = strings.Replace(templateDataStr, "{{TARGET_COMPILE_DEFINITIONS}}",
        strings.Join(targetInfo.Definitions.GetTargetDefinitions(), " "), -1)

    if !utils.PathExists(appPath + io.Sep + ".wio" + io.Sep + "build") {
        err := os.MkdirAll(appPath+io.Sep+".wio"+io.Sep+"build", os.ModePerm)
        if err != nil {
            return err
        }
    }

    return io.NormalIO.WriteFile(appPath+io.Sep+".wio"+io.Sep+"build"+io.Sep+"CMakeLists.txt",
        []byte(templateDataStr))
}
