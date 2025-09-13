<div align="center">

<img src=".github/resources/mortar-logo.png" width="auto" alt="Mortar wordmark">
<h3 style="font-size: 25px;">
    A ROM download client for <a href="https://nextui.loveretro.games" target="_blank">NextUI</a>
</h3>

<h4 style="font-size: 18px;">
Supports RomM and Megathread

Art Downloads powered by the _Libretro Thumbnail Project_
</h4>

## [Download this in Pak Store!](https://github.com/UncleJunVIP/nextui-pak-store)

![GitHub License](https://img.shields.io/github/license/UncleJunVip/Mortar?style=for-the-badge&color=007C77)
![GitHub Release](https://img.shields.io/github/v/release/UncleJunVIP/Mortar?sort=semver&style=for-the-badge&color=007C77)
![GitHub Repo stars](https://img.shields.io/github/stars/UncleJunVip/Mortar?style=for-the-badge&color=007C77)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/UncleJunVIP/Mortar/total?style=for-the-badge&label=Total%20Downloads&color=007C77)


</div>

---

## How do I setup Mortar?

1. Own a TrimUI Brick or Smart Pro and have a SD Card with NextUI.
2. Connect your device to a Wi-Fi network.
3. Download the latest Mortar release from this repo or install it using [Pak Store](https://github.com/UncleJunVIP/nextui-pak-store).
   - If downloading manually, unzip the release before continuing.
4. Edit one of the templates [found here](/.github/resources/config_examples).
   - Use your 🧠. I believe in you!
5. Save the edited template as `config.yml`.
    - Pak Store Users: upload `config.yml` to `SD_ROOT/Tools/tg5040/Mortar.pak`.
    - Manual Installers: upload the `Mortar.pak` directory that came in the releases zip and place `config.yml` inside.
6. Launch `Mortar` from the `Tools` menu and enjoy!

---

## Configuration Reference

```yaml
hosts:
  - display_name: "Display Name"
    host_type: ROMM # Valid Choices: ROMM | MEGATHREAD
    root_uri: "https://domain.tld" # This can be the start of a URL with protocol (e.g. https://), a host name or an IP Address

    port: 445 # Optional otherwise unless using non-standard ports

    username: "GUEST" # Used by RomM
    password: "hunter2" # Used by RomM

    platforms: # One or more mappings of the host directory to the local filesystem
      - platform_name: "Game Boy" # Name it whatever you want
        system_tag: "GB" # Must match the tag in the `SDCARD_ROOT/Roms` directories
        local_directory: "/mnt/SDCARD/Roms/Game Boy (GB)/" # Explicitly set the path. This will be overwritten if `system_tag` is set
        host_subdirectory: "/files/No-Intro/Nintendo%20-%20Game%20Boy/" # The subdirectory on the host, not used by RomM
        romm_platform_id: "1" # Used by RomM in place of `host_subdirectory`
        skip_inclusive_filters: false # If true, everything in the host directory will be included
        skip_exclusive_filters: false # If true, nothing in the host directory will be excluded
        is_arcade: false # If true, Mortar will use an internal mapping file for arcade names

        # Define more sections if desired

    filters:
      inclusive_filters: # Inclusive filters are applied first. If the ROM filename contains any of these, it will be included
        - "USA"
        - "En,"
      exclusive_filters: # Exclusive filters are applied second. If the ROM filename contains any of these, it will be excluded
        - "(Proto"
        - "(Demo)"
        - "(Beta)"
        - "(Aftermarket"
        - "4-in-1"
        - "4 in 1"
        - "(Europe)"
        - "(Japan)"

  # Define more hosts if desired

download_art: true # If true, Mortar will attempt to find box art. If found, it will display it and let you indicate if you want it
art_download_type: "BOX_ART" # Optional, defaults to BOX_ART. Does not impact art downloads from RoMM. Valid Choices: BOX_ART | TITLE_SCREEN | LOGOS | SCREENSHOTS
log_level: "ERROR" # Optional, defaults to error. Handy when shit breaks
```

Sample configuration files can be [found here](/.github/resources/config_examples).

---

## Enjoying Mortar?

You might be interested in my other NextUI Paks!

[Pak Store](https://github.com/UncleJunVIP/nextui-pak-store) - install, update and manage the amazing work from the community right on device

[Game Manager](https://github.com/UncleJunVIP/nextui-game-manager) - manage your ROM library right on device

---

## Be a friend, tell a friend something nice; it might change their life!

I've spent a good chunk of time building Mortar.

If you feel inclined to pay it forward, go do something nice for someone! ❤️

✌🏻
