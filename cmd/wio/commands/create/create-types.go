package create

import "github.com/urfave/cli"

type Create struct {
    Context *cli.Context
    Update  bool
    error   error
}

type PlatformInfo struct {
    Frameworks []string
    Boards     []string
}

type createInfo struct {
    Directory string
    Type      string
    Name      string

    Platform              map[string]*PlatformInfo
    PlatformNamesSimple   []string
    PlatformNamesDetailed []string
    BoardNamesSimple      []string
    BoardNamesDetailed    []string

    ConfigOnly bool
    HeaderOnly bool
}

// get context for the command
func (create Create) GetContext() *cli.Context {
    return create.Context
}

// Executes the create command
func (create Create) Execute() {
    directory := performDirectoryCheck(create.Context)

    if create.Update {
        // this checks if wio.yml file exists for it to update
        performWioExistsCheck(directory)
        // this checks if project is valid state to be updated
        performPreUpdateCheck(directory, &create)
        create.handleUpdate(directory)
    } else {
        // this checks if directory is empty before create can be triggered
        performPreCreateCheck(directory, create.Context.Bool("only-config"))
        create.createPackageProject(directory)
    }
}
