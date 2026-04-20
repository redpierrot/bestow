package cmd

const RootCmdShort string = "Modern symlink manager for managing dotfiles and configs with ease"
const RootCmdLong string = `A modern and fast symlink manager for managing dotfiles and configurations.
Bestow is the spiritual successor of the GNU stow.

Bestow will create and manage symlinks between a given source and destination
directory.

Each directory inside the source is considered a package. Each package can
contain multiple files. When creating the symlinks, bestow honors the directory
structure inside the package.

Bestow handles conflicts before execution, and provides multiple options and
alternatives for handling conflicts.
`
const RootCmdExamples string = `# Initialize the configs
bestow init

# Stow all the packages and files in the source
bestow stow

# Stow only the 'nvim' package
bestow stow nvim

# Provide custom source and destination
bestow stow nvim --source ~/dotfiles --destination ~/

# Backup files if there's a conflict
bestow stow git --backup

# Unstow all the packages
bestow unstow

# Unstow specific package
bestow unstow nvim

# Dry run before execution
bestow unstow nvim --dry-run
`

const InitShort string = "Initializes the bestow configurations"
const InitLong string = `The init command will initialize the bestow configurations.
If the XDG_CONFIG_HOME is set, it will be treated as the config home directory.

If it is not set, the default "$HOME/.config" directory will be used as the config home.
`

const StowShort string = "Stow creates symlinks between the source packages to destination"
const StowLong string = `Stow creates symlinks between the source and the destination.

Each directory inside the source is considered a package. When creating
symlinks, bestow keeps the directory structure inside the package.

$source/zsh/.zshrc -> $destination/.zshrc
$source/nvim/.config/nvim/init.lus -> $destination/.config/nvim/init.lua

Bestow does not create symlinks for the subdirectories, instead it creates
directories and links only the files.

If no packages are provided, all the packages will be stowed.
`
const StowExamples string = `# Stow only given packages
bestow stow nvim git

# Stow while replacing any existing file forcefully
bestow stow --force
`

const UnstowShort string = "Removes the stowed symlinks from the destination"
const UnstowLong string = `Removes the symlinks created by bestow.

Unstow walks the source package tree and removes the corresponding symlinks
from the destination.

If no packages are provided, all the packages will be unstowed.
`
const UnstowExamples string = `# Unstow only given packages
bestow unstow nvim git
`
