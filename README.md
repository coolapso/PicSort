<p align="center">
  <img src="https://github.com/coolapso/picsort/blob/dev/media/logo.png" width="200" >
</p>

# picsort
[![Release](https://github.com/coolapso/picsort/actions/workflows/release.yaml/badge.svg?branch=main)](https://github.com/coolapso/picsort/actions/workflows/release.yaml)
![GitHub Tag](https://img.shields.io/github/v/tag/coolapso/picsort?logo=semver&label=semver&labelColor=gray&color=green)
[![Go Report Card](https://goreportcard.com/badge/github.com/coolapso/picsort)](https://goreportcard.com/report/github.com/coolapso/picsort)
![GitHub Sponsors](https://img.shields.io/github/sponsors/coolapso?style=flat&logo=githubsponsors)

keyboard-driven tool for sorting images into folders.

Picsort is a desktop application designed to help you rapidly organize large sets of pictures into different folders, primarily using your keyboard. While it was created to assist with sorting image datasets for training computer vision models, it's versatile enough for any large-scale photo organization task.

It is important to clarify that Picsort is not a replacement for general-purpose gallery or photo management tools. Its sole mission is to make the sorting process as fast and ergonomic as possible.

### Features

*   **Keyboard-First Design**: Navigate, select, and sort images without leaving the keyboard.
*   **Vim-like Keybindings**: Use `HJKL` keys for efficient navigation.
*   **High Performance**: Picsort tries to be as fast as possible, even with thousands of images, however the tradeoff is that it will take a while to load the first time while picsort generates a cache with thumbnails and previews for a fast and smooth experience.
*   **Non-Destructive**: Your original images are sacred. Picsort never modifies them. It reads them to create a cache and copies them to your chosen destination upon export.
*   **Simple UI**: The UI aims to be as simple ans self explanatory as possible without much going on!


## How to use picsort

You can watch a quick demo of picsort [here](https://youtu.be/HdG0HuAClu0)

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
