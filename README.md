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

## Prepare required configuration file:
- The conf file should be written as below.
- The conf file path : '~/.config/go-telegram-bot/go-telegram-bot.conf'
```
[telegram]
token_key		# Telegram bot API token
admin_user		# Admin username
downloads_dir		# Download dir path from chat

[qbittorrent]
url			# qBittorrent website url address
username		# qBittorrent username
password		# qBittorrent user password
downloads_dir		# Torrent seed download dir path

[watch]
clipboard_file		# Watchied file path, Save text path
downloads_dir		# Watchied dir path, Completed torrent
```
- e.g. go-telegram-bot.conf
```
[telegram]
token_key =
admin_user =
downloads_dir =

[qbittorrent]
url =
username =
password =
torrent_dir =

[watch]
clipboard_file =
downloads_dir =
```
