<p align="center">
  <img src="https://github.com/coolapso/picsort/blob/main/media/logo.png" width="200" >
</p>

# picsort
[![Release](https://github.com/coolapso/picsort/actions/workflows/release.yaml/badge.svg?branch=main)](https://github.com/coolapso/picsort/actions/workflows/release.yaml)
![GitHub Tag](https://img.shields.io/github/v/tag/coolapso/picsort?logo=semver&label=semver&labelColor=gray&color=green)
[![Go Report Card](https://goreportcard.com/badge/github.com/coolapso/picsort)](https://goreportcard.com/report/github.com/coolapso/picsort)
![GitHub Sponsors](https://img.shields.io/github/sponsors/coolapso?style=flat&logo=githubsponsors)

keyboard-driven tool for sorting images into folders.

Picsort is a desktop application designed to help you rapidly organize and classify large sets of pictures, primarily using your keyboard. While it was created to assist with sorting image datasets for training computer vision models, it's versatile enough for any large-scale photo organization task.

It is important to clarify that Picsort is not a replacement for general-purpose gallery or photo management tools. Its sole mission is to make the sorting process as fast and ergonomic as possible.

### Features

*   **Keyboard-First Design**: Navigate, select, and sort images without leaving the keyboard.
*   **Vim-like Keybindings**: Use `HJKL` keys for efficient navigation.
*   **High Performance**: Picsort tries to be as fast as possible, even with thousands of images, however the tradeoff is that it will take a while to load the first time while picsort generates a cache with thumbnails and previews for a fast and smooth experience.
*   **Non-Destructive**: Your original images are sacred. Picsort never modifies them. It reads them to create a cache and copies them to your chosen destination upon export.
*   **Simple UI**: The UI aims to be as simple ans self explanatory as possible without much going on!

### Why I created Picsort?

I live in the Arctic Circle, where the Northern Lights are frequently visible. I set up a 24/7 live stream of the sky, both as an early warning system for when the aurora is active and to share the view with everyone.

As part of this project, I wanted to connect a computer vision model to the video feed to automatically detect auroral activity and send notifications. However, I quickly ran into two problems:

1.  There were no pre-trained models for this, which meant I had to train my own.
2.  Training a model requires a large, well-organized dataset of images. Getting the dataset right involves frequent sorting, tweaking, and updating.

I tried to manage my images with traditional file explorers, Darktable, and many other tools, but none of them felt efficient enough for the task. This led me down a bit of a yak-shaving journey, and Picsort was born!

if you want to check out the live stream you can find it both on [youtube](https://youtube.com/@thearcticskies) and on [twitch](https://twitch.tv/thearcticskies)

## How to use picsort

You can watch a quick demo of picsort [here](https://youtu.be/HdG0HuAClu0)

[![Demo](https://img.youtube.com/vi/HdG0HuAClu0/0.jpg)](https://www.youtube.com/watch?v=HdG0HuAClu0)

### How it works

when you open a dataset, picsort will generate a cache with thumbnails and previews, this task is multi threaded and uses all available cores, once the cache is generated, the subsequent loads of your dataset should be significantly faster. All operations are then done using the cache, and the original images are never touched. Upen exporting picsort will copy the images from the original location to the chosen destionation and the images will be placed in a directory wit hthe correspoinding number.

### Keyboard Shortcuts

At any time, press `?` to view the help menu with all available keybindings.

| Action                 | Shortcut                       |
| ---------------------- | ------------------------------ |
| Open Dataset           | `Ctrl+O`                       |
| Export Sorted Images   | `Ctrl+E`                       |
| Navigate Thumbnails    | `H`, `J`, `K`, `L` / Arrow Keys|
| Switch Bin/Tab         | `Ctrl` + `0-9`                 |
| Move to Bin            | `1-9`                          |
| Move to Unsorted       | `0`                            |
| Select/Deselect Image  | `Space`                        |
| Multi-select Range     | `Shift` + Navigation Key       |
| Add Bin                | `Ctrl+T`                       |
| Remove Bin             | `Ctrl+W`                       |
| Resize Panels          | `Ctrl+H` / `Ctrl+L`            |
| Show Help Menu         | `?` or `F1`                    |

Thank you for checking out Picsort. I hope you find it useful!

### How to install

Picsort is simple to install and there are a few ways to do it, more ways to install can be added in the future if there's interest for it, pease let me know!

#### Debian based distros

Grab the debian package from the [releases page](https://github.com/coolapso/PicSort/releases) and install it with `sudo apt install ./picsort_1.1.0_amd64.deb`

#### RPM based distros

Grab the rpm package from the [releases page](https://github.com/coolapso/PicSort/releases) and install it with `sudo dnf install picsort-1.1.0-1.x86_64.rpm`

#### Arch based distros (AUR)

There's a arch linux AUR package available: 

```bash
yay -S picsort-bin
```

#### Install script

> [!WARNING] 
> Please note that curl to bash is not the most secure way to install any project. Please make sure you understand and trust the [install script](https://github.com/coolapso/picsort/blob/main/build/install.sh) before running it.

**Latest version**
```bash
curl -L https://coolapso.github.io/PicSort/install.sh | bash
```

**Specific version**
```bash
curl -L https://coolapso.github.io/PicSort/install.sh | VERSION="v1.1.0" bash
```
#### Manually

Picsort is just a binary, and ther's actually no real need for instalation. 
all the AUR package and install script do is to place picsort in your path and adding a icon and a .desktop entry, but if you want you can just grab the binary from the [releases page](https://github.com/coolapso/PicSort/releases) and just run it

```bash
VERSION=$(curl -s "https://api.github.com/repos/coolapso/picsort/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
curl -LO https://github.com/coolapso/picsort/releases/download/$VERSION/PicSort_"${VERSION:1}"_linux_amd64.tar.gz
tar -xzf PicSort_${VERSION:1}_linux_amd64.tar.gz picsort
./picsrot
```

#### Go Package

it is also possible to install it with go 

**Latest version**
```bash
go install github.com/coolapso/picsort
```

**Specific version**
```bash
go install github.com/coolapso/picsort@v0.1.3
```

### How to uninstall

To uninstall you just uinstall with your package manager, ir if you used the install script you can use the uninstall script: 

```bash
curl -L https://coolapso.github.io/PicSort/uninstall.sh | bash
```

> [!WARNING] 
> Again, plaese make sure you understand and trust the [uninstall script](https://github.com/coolapso/picsort/blob/main/build/uninstall.sh) before running it. The script is pretty simple you can just run the commands yourself!


# Contributions

Improvements and suggestions are always welcome, feel free to check for any open issues, open a new Issue or Pull Request

If you like this project and want to support / contribute in a different way you can always [:heart: Sponsor Me](https://github.com/sponsors/coolapso) or

<a href="https://www.buymeacoffee.com/coolapso" target="_blank">
  <img src="https://cdn.buymeacoffee.com/buttons/default-yellow.png" alt="Buy Me A Coffee" style="height: 51px !important;width: 217px !important;" />
</a>
