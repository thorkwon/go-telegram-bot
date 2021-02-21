# Go Telegram bot

## Features
- Monitoring a specific file and then send the file contents to the telegram chat.
- Download the file received from admin user private chat to localhost.
- Save text msg to localhost from chat.
- Notification for COVID-19 status update. (Send to chat once a day)
- Download torrent seed from admin user chat, have to specify the path.
- Notification for torrent download complete, and delete completed seed.

## Install options packages:
- If someone wants the automatic deletion feature for completed seed.
  Chromedriver installation is required to use this feature.
```
apt install chromium-browser chromium-codecs-ffmpeg chromium-chromedriver
```

## Prepare required files:
- The below files have to in the '~/.config/go-telegram-bot' directory.
```
token_key		# Telegram bot API token
watch_file		# Watchied file path, Save text path
download_dir		# Download dir path
admin_user		# Admin username
torrent_dir		# Torrent seed download dir path
qbit_downloads_dir	# Watchied dir path, Completed torrent
qbit_url		# qBittorrent website url address
qbit_username		# qBittorrent username
qbit_password		# qBittorrent user password
```
