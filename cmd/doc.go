/*
All Rights Reversed (ɔ)
*/

package cmd

const rootCmdShort = "Modern symlink manager for managing dotfiles and configs with ease"
const rootCmdLong = `A modern and fast symlink manager for managing dotfiles and configurations.
Bestow is the spiritual successor of the GNU stow.

Bestow will create and manage symlinks between a given source and destination
directory.

Each directory inside the source is considered a package. Each package can
contain multiple files. When creating the symlinks, bestow honors the directory
structure inside the package.

Bestow handles conflicts before execution, and provides multiple options and
alternatives for handling conflicts.
`
const rootCmdExamples = `# Initialize the configs
bestow init

# Stow all the packages and files in the source
bestow stow

# Stow only the 'nvim' package
bestow stow nvim

# Provide custom source and destination
bestow stow nvim --source ~/dotfiles/directory --destination ~/sandbox/directory

# Backup files if there's a conflict
bestow stow git --backup

# Unstow all the packages
bestow unstow

# Unstow specific package
bestow unstow nvim

# Dry run before execution
bestow unstow nvim --dry-run
`

const initShort = "Initializes the bestow configurations"
const initLong = `The init command will initialize the bestow configurations.
If the XDG_CONFIG_HOME is set, it will be treated as the config home directory.

If it is not set, the default "$HOME/.config" directory will be used as the config home.
`
const initExamples = `# Init with providing source
bestow init --source ~/dotfiles

# Initialize with source and destination
bestow init --source ~/dotfiles --destination ~/sandbox

# Provide custom ignore list
bestow init --source ~/dotfiles --destination ~/sandbox --ignore-list ".git, node_modules, README.md"
`

const stowShort = "Stow creates symlinks between the source packages to destination"
const stowLong = `Stow creates symlinks between the source and the destination.

Each directory inside the source is considered a package. When creating
symlinks, bestow keeps the directory structure inside the package.

$source/zsh/.zshrc -> $destination/.zshrc
$source/nvim/.config/nvim/init.lua -> $destination/.config/nvim/init.lua

Bestow does not create symlinks for the subdirectories, instead it creates
directories and links only the files.

If no packages are provided, all the packages will be stowed.
`
const stowExamples = `# Stow only given packages
bestow stow nvim git

# Stow while replacing any existing file forcefully
bestow stow --force
`

const unstowShort = "Removes the stowed symlinks from the destination"
const unstowLong = `Removes the symlinks created by bestow.

Unstow walks the source package tree and removes the corresponding symlinks
from the destination.

If no packages are provided, all the packages will be unstowed.

The unstow command will remove any empty directories it encounters.
You can configure bestow to not remove empty directories while unstowing.
`
const unstowExamples = `# Unstow only given packages
bestow unstow nvim git

# Unstow all the packages
bestow unstow

# Provide custom source and destination
bestow unstow -s ~/dotfiles -d ~/sandbox
`
